package services

import (
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"cinema/internal/repository"
	"encoding/json"
	"errors"
	"strings"

	appErrors "cinema/internal/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	telegramNotificationTypeWeeklyPackResults = "weekly_pack_results"
	telegramNotificationTypePendingUser       = "pending_user"
	telegramNotificationTypeMovieProposal     = "movie_proposal"

	telegramNotificationStatusQueued = "queued"
	telegramNotificationStatusSent   = "sent"
	telegramNotificationStatusFailed = "failed"
)

type TelegramNotificationService interface {
	List(status, notificationType string) ([]dto.TelegramNotificationView, error)
	Enqueue(enqueuedByUserID uint, req dto.EnqueueTelegramNotificationRequest) (dto.TelegramNotificationView, error)
	Retry(notificationID, adminID uint) (dto.TelegramNotificationView, error)
}

type telegramNotificationService struct {
	rep repository.Repository
}

func NewTelegramNotificationService(rep repository.Repository) TelegramNotificationService {
	return &telegramNotificationService{rep: rep}
}

func (s *telegramNotificationService) List(status, notificationType string) ([]dto.TelegramNotificationView, error) {
	query := s.rep.Preload("TargetUser").Preload("EnqueuedBy").Order("created_at DESC, id DESC")
	if strings.TrimSpace(status) != "" {
		normalized, err := normalizeTelegramNotificationStatus(status)
		if err != nil {
			return nil, err
		}
		query = query.Where("status = ?", normalized)
	}
	if strings.TrimSpace(notificationType) != "" {
		normalized, err := normalizeTelegramNotificationType(notificationType)
		if err != nil {
			return nil, err
		}
		query = query.Where("type = ?", normalized)
	}

	var notifications []models.TelegramNotification
	if err := query.Find(&notifications).Error; err != nil {
		return nil, err
	}
	return mapTelegramNotificationViews(notifications), nil
}

func (s *telegramNotificationService) Enqueue(enqueuedByUserID uint, req dto.EnqueueTelegramNotificationRequest) (dto.TelegramNotificationView, error) {
	notificationType, err := normalizeTelegramNotificationType(req.Type)
	if err != nil {
		return dto.TelegramNotificationView{}, err
	}
	if req.EntityID == 0 {
		return dto.TelegramNotificationView{}, appErrors.ErrInvalidTelegramNotificationPayload
	}
	if err := s.validateNotificationEntity(notificationType, req.EntityID); err != nil {
		return dto.TelegramNotificationView{}, err
	}
	if req.TargetUserID != nil {
		if err := s.ensureUserExists(*req.TargetUserID); err != nil {
			return dto.TelegramNotificationView{}, err
		}
	}

	payload, err := buildTelegramNotificationPayload(notificationType, req.EntityID, req.Payload)
	if err != nil {
		return dto.TelegramNotificationView{}, err
	}

	notification := models.TelegramNotification{
		Type:             notificationType,
		Status:           telegramNotificationStatusQueued,
		TargetUserID:     req.TargetUserID,
		EntityID:         req.EntityID,
		Payload:          payload,
		EnqueuedByUserID: enqueuedByUserID,
	}
	if err := s.rep.Create(&notification).Error; err != nil {
		return dto.TelegramNotificationView{}, err
	}

	return s.GetByID(notification.ID)
}

func (s *telegramNotificationService) Retry(notificationID, adminID uint) (dto.TelegramNotificationView, error) {
	err := s.rep.Transaction(func(tx repository.Repository) error {
		var notification models.TelegramNotification
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&notification, notificationID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrTelegramNotificationNotFound
			}
			return err
		}
		if notification.Status != telegramNotificationStatusFailed {
			return appErrors.ErrTelegramNotificationNotRetryable
		}

		updates := map[string]interface{}{
			"status":     telegramNotificationStatusQueued,
			"last_error": "",
			"sent_at":    nil,
		}
		return tx.Model(&models.TelegramNotification{}).Where("id = ?", notificationID).Updates(updates).Error
	})
	if err != nil {
		return dto.TelegramNotificationView{}, err
	}
	return s.GetByID(notificationID)
}

func (s *telegramNotificationService) GetByID(id uint) (dto.TelegramNotificationView, error) {
	var notification models.TelegramNotification
	if err := s.rep.Preload("TargetUser").Preload("EnqueuedBy").First(&notification, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.TelegramNotificationView{}, appErrors.ErrTelegramNotificationNotFound
		}
		return dto.TelegramNotificationView{}, err
	}
	return mapTelegramNotificationView(notification), nil
}

func (s *telegramNotificationService) validateNotificationEntity(notificationType string, entityID uint) error {
	switch notificationType {
	case telegramNotificationTypeWeeklyPackResults:
		var pack models.WeeklyPack
		if err := s.rep.First(&pack, entityID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrWeeklyPackNotFound
			}
			return err
		}
		if pack.Status != weeklyPackStatusClosed && pack.Status != weeklyPackStatusArchived {
			return appErrors.ErrInvalidTelegramNotificationEntity
		}
	case telegramNotificationTypePendingUser:
		var user models.User
		if err := s.rep.First(&user, entityID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrUserNotFound
			}
			return err
		}
		if user.Status != "pending" {
			return appErrors.ErrInvalidTelegramNotificationEntity
		}
	case telegramNotificationTypeMovieProposal:
		var proposal models.MovieProposal
		if err := s.rep.First(&proposal, entityID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrMovieProposalNotFound
			}
			return err
		}
		if proposal.Status != proposalStatusPending {
			return appErrors.ErrInvalidTelegramNotificationEntity
		}
	default:
		return appErrors.ErrInvalidTelegramNotificationType
	}
	return nil
}

func (s *telegramNotificationService) ensureUserExists(userID uint) error {
	var user models.User
	if err := s.rep.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrUserNotFound
		}
		return err
	}
	return nil
}

func normalizeTelegramNotificationType(notificationType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(notificationType)) {
	case telegramNotificationTypeWeeklyPackResults:
		return telegramNotificationTypeWeeklyPackResults, nil
	case telegramNotificationTypePendingUser:
		return telegramNotificationTypePendingUser, nil
	case telegramNotificationTypeMovieProposal:
		return telegramNotificationTypeMovieProposal, nil
	default:
		return "", appErrors.ErrInvalidTelegramNotificationType
	}
}

func normalizeTelegramNotificationStatus(status string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case telegramNotificationStatusQueued:
		return telegramNotificationStatusQueued, nil
	case telegramNotificationStatusSent:
		return telegramNotificationStatusSent, nil
	case telegramNotificationStatusFailed:
		return telegramNotificationStatusFailed, nil
	default:
		return "", appErrors.ErrInvalidTelegramNotificationStatus
	}
}

func buildTelegramNotificationPayload(notificationType string, entityID uint, raw json.RawMessage) (string, error) {
	if len(raw) > 0 {
		if !json.Valid(raw) {
			return "", appErrors.ErrInvalidTelegramNotificationPayload
		}
		var payload map[string]interface{}
		if err := json.Unmarshal(raw, &payload); err != nil || payload == nil {
			return "", appErrors.ErrInvalidTelegramNotificationPayload
		}
		return string(raw), nil
	}

	payload := map[string]uint{}
	switch notificationType {
	case telegramNotificationTypeWeeklyPackResults:
		payload["weekly_pack_id"] = entityID
	case telegramNotificationTypePendingUser:
		payload["user_id"] = entityID
	case telegramNotificationTypeMovieProposal:
		payload["proposal_id"] = entityID
	default:
		return "", appErrors.ErrInvalidTelegramNotificationType
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func mapTelegramNotificationViews(notifications []models.TelegramNotification) []dto.TelegramNotificationView {
	out := make([]dto.TelegramNotificationView, 0, len(notifications))
	for _, notification := range notifications {
		out = append(out, mapTelegramNotificationView(notification))
	}
	return out
}

func mapTelegramNotificationView(notification models.TelegramNotification) dto.TelegramNotificationView {
	payload := json.RawMessage(notification.Payload)
	if !json.Valid(payload) {
		payload = json.RawMessage(`{}`)
	}
	return dto.TelegramNotificationView{
		ID:                 notification.ID,
		Type:               notification.Type,
		Status:             notification.Status,
		TargetUserID:       notification.TargetUserID,
		TargetUsername:     notification.TargetUser.Username,
		TargetDisplayName:  notification.TargetUser.DisplayName,
		EntityID:           notification.EntityID,
		Payload:            payload,
		Attempts:           notification.Attempts,
		LastError:          notification.LastError,
		EnqueuedByUserID:   notification.EnqueuedByUserID,
		EnqueuedByUsername: notification.EnqueuedBy.Username,
		EnqueuedByName:     notification.EnqueuedBy.DisplayName,
		SentAt:             notification.SentAt,
		CreatedAt:          notification.CreatedAt,
		UpdatedAt:          notification.UpdatedAt,
	}
}
