package services

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"cinema/internal/repository"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	proposalStatusPending  = "pending"
	proposalStatusApproved = "approved"
	proposalStatusRejected = "rejected"

	proposalSourceManual    = "manual"
	proposalSourceKinopoisk = "kinopoisk"
)

type MovieProposalService interface {
	Create(userID uint, req dto.CreateMovieProposalRequest) (dto.MovieProposalView, error)
	List(status string) ([]dto.MovieProposalView, error)
	ListByUser(userID uint) ([]dto.MovieProposalView, error)
	GetByID(id uint) (dto.MovieProposalView, error)
	Moderate(proposalID, adminID uint, req dto.ModerateMovieProposalRequest) (dto.MovieProposalView, error)
}

type movieProposalService struct {
	rep repository.Repository
}

func NewMovieProposalService(rep repository.Repository) MovieProposalService {
	return &movieProposalService{rep: rep}
}

func (s *movieProposalService) Create(userID uint, req dto.CreateMovieProposalRequest) (dto.MovieProposalView, error) {
	proposal, err := buildMovieProposal(userID, req)
	if err != nil {
		return dto.MovieProposalView{}, err
	}

	if err := s.rep.Create(&proposal).Error; err != nil {
		return dto.MovieProposalView{}, err
	}

	return s.GetByID(proposal.ID)
}

func (s *movieProposalService) List(status string) ([]dto.MovieProposalView, error) {
	query := s.rep.Preload("ProposedBy").Preload("ModeratedBy").Order("created_at DESC, id DESC")
	if strings.TrimSpace(status) != "" {
		normalized, err := normalizeProposalStatus(status, true)
		if err != nil {
			return nil, err
		}
		query = query.Where("status = ?", normalized)
	}

	var proposals []models.MovieProposal
	if err := query.Find(&proposals).Error; err != nil {
		return nil, err
	}
	return mapMovieProposalViews(proposals), nil
}

func (s *movieProposalService) ListByUser(userID uint) ([]dto.MovieProposalView, error) {
	var proposals []models.MovieProposal
	if err := s.rep.Preload("ProposedBy").Preload("ModeratedBy").
		Where("proposed_by_user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Find(&proposals).Error; err != nil {
		return nil, err
	}
	return mapMovieProposalViews(proposals), nil
}

func (s *movieProposalService) Moderate(proposalID, adminID uint, req dto.ModerateMovieProposalRequest) (dto.MovieProposalView, error) {
	status, err := normalizeProposalStatus(req.Status, false)
	if err != nil {
		return dto.MovieProposalView{}, err
	}
	if status == proposalStatusPending {
		return dto.MovieProposalView{}, appErrors.ErrInvalidMovieProposalStatus
	}

	err = s.rep.Transaction(func(tx repository.Repository) error {
		var proposal models.MovieProposal
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&proposal, proposalID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrMovieProposalNotFound
			}
			return err
		}
		if proposal.Status != proposalStatusPending {
			return appErrors.ErrMovieProposalAlreadyClosed
		}

		now := time.Now()
		proposal.Status = status
		proposal.ModeratedByUserID = &adminID
		proposal.ModeratedAt = &now
		proposal.ModerationComment = strings.TrimSpace(req.ModerationComment)

		if status == proposalStatusApproved {
			film, err := findOrCreateFilmFromProposal(tx, proposal)
			if err != nil {
				return err
			}
			proposal.FilmID = &film.ID
		}

		return tx.Save(&proposal).Error
	})
	if err != nil {
		return dto.MovieProposalView{}, err
	}

	return s.GetByID(proposalID)
}

func (s *movieProposalService) GetByID(id uint) (dto.MovieProposalView, error) {
	var proposal models.MovieProposal
	if err := s.rep.Preload("ProposedBy").Preload("ModeratedBy").First(&proposal, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.MovieProposalView{}, appErrors.ErrMovieProposalNotFound
		}
		return dto.MovieProposalView{}, err
	}
	return mapMovieProposalView(proposal), nil
}

func buildMovieProposal(userID uint, req dto.CreateMovieProposalRequest) (models.MovieProposal, error) {
	title := strings.TrimSpace(req.Title)
	description := strings.TrimSpace(req.Description)
	smallDescription := strings.TrimSpace(req.SmallDescription)
	country := strings.TrimSpace(req.Country)
	poster := strings.TrimSpace(req.Poster)
	source, err := normalizeProposalSource(req.Source)
	if err != nil {
		return models.MovieProposal{}, err
	}

	if title == "" || description == "" || smallDescription == "" || country == "" || poster == "" ||
		req.Duration <= 0 || req.ReleaseDate <= 0 || req.RatingKp < 0 || req.RatingKp > 10 {
		return models.MovieProposal{}, appErrors.ErrInvalidMovieProposalPayload
	}

	return models.MovieProposal{
		Title:            title,
		Description:      description,
		SmallDescription: smallDescription,
		Duration:         req.Duration,
		ReleaseDate:      req.ReleaseDate,
		Country:          country,
		Poster:           poster,
		RatingKp:         req.RatingKp,
		Source:           source,
		Status:           proposalStatusPending,
		ProposedByUserID: userID,
	}, nil
}

func normalizeProposalSource(source string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "", proposalSourceManual:
		return proposalSourceManual, nil
	case proposalSourceKinopoisk:
		return proposalSourceKinopoisk, nil
	default:
		return "", appErrors.ErrInvalidMovieProposalPayload
	}
}

func normalizeProposalStatus(status string, allowPending bool) (string, error) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case proposalStatusPending:
		if allowPending {
			return proposalStatusPending, nil
		}
	case proposalStatusApproved:
		return proposalStatusApproved, nil
	case proposalStatusRejected:
		return proposalStatusRejected, nil
	}
	return "", appErrors.ErrInvalidMovieProposalStatus
}

func findOrCreateFilmFromProposal(rep repository.Repository, proposal models.MovieProposal) (models.Film, error) {
	var film models.Film
	if err := rep.Where("LOWER(title) = LOWER(?)", proposal.Title).First(&film).Error; err == nil {
		if film.ReleaseDate == proposal.ReleaseDate && strings.EqualFold(strings.TrimSpace(film.Country), strings.TrimSpace(proposal.Country)) {
			return film, nil
		}
		return models.Film{}, appErrors.ErrMovieProposalDuplicateFilm
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Film{}, err
	}

	film = models.Film{
		Title:            proposal.Title,
		Description:      proposal.Description,
		SmallDescription: proposal.SmallDescription,
		Duration:         proposal.Duration,
		ReleaseDate:      proposal.ReleaseDate,
		Country:          proposal.Country,
		Poster:           proposal.Poster,
		RatingKp:         proposal.RatingKp,
	}
	if err := rep.Create(&film).Error; err != nil {
		if isUniqueConstraintError(err) {
			return models.Film{}, appErrors.ErrMovieProposalDuplicateFilm
		}
		return models.Film{}, err
	}
	return film, nil
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate")
}

func mapMovieProposalViews(proposals []models.MovieProposal) []dto.MovieProposalView {
	out := make([]dto.MovieProposalView, 0, len(proposals))
	for _, proposal := range proposals {
		out = append(out, mapMovieProposalView(proposal))
	}
	return out
}

func mapMovieProposalView(proposal models.MovieProposal) dto.MovieProposalView {
	return dto.MovieProposalView{
		ID:                  proposal.ID,
		Title:               proposal.Title,
		Description:         proposal.Description,
		SmallDescription:    proposal.SmallDescription,
		Duration:            proposal.Duration,
		ReleaseDate:         proposal.ReleaseDate,
		Country:             proposal.Country,
		Poster:              proposal.Poster,
		RatingKp:            proposal.RatingKp,
		Source:              proposal.Source,
		Status:              proposal.Status,
		ProposedByUserID:    proposal.ProposedByUserID,
		ProposedByUsername:  proposal.ProposedBy.Username,
		ProposedByName:      proposal.ProposedBy.DisplayName,
		ModeratedByUserID:   proposal.ModeratedByUserID,
		ModeratedByUsername: proposal.ModeratedBy.Username,
		ModeratedByName:     proposal.ModeratedBy.DisplayName,
		ModeratedAt:         proposal.ModeratedAt,
		ModerationComment:   proposal.ModerationComment,
		FilmID:              proposal.FilmID,
		CreatedAt:           proposal.CreatedAt,
		UpdatedAt:           proposal.UpdatedAt,
	}
}
