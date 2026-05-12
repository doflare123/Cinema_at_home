package services

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"cinema/internal/repository"
	"errors"
	"strings"

	"gorm.io/gorm"
)

const (
	expectationTargetMovie     = "movie"
	expectationTargetFranchise = "franchise"
	expectationVoteScore       = "score"
	expectationVoteRefuse      = "refuse"
)

type ExpectationService interface {
	Upsert(userID uint, req dto.UpsertExpectationRequest) (dto.ExpectationVoteView, error)
	Summary(targetType string, targetID uint) (dto.ExpectationSummaryView, error)
	ListByUser(userID uint) ([]dto.ExpectationVoteView, error)
}

type expectationService struct {
	rep repository.Repository
}

func NewExpectationService(rep repository.Repository) ExpectationService {
	return &expectationService{rep: rep}
}

func (s *expectationService) Upsert(userID uint, req dto.UpsertExpectationRequest) (dto.ExpectationVoteView, error) {
	targetType := strings.TrimSpace(req.TargetType)
	voteType := strings.TrimSpace(req.VoteType)
	if err := validateExpectationRequest(req, targetType, voteType); err != nil {
		return dto.ExpectationVoteView{}, err
	}

	if err := s.ensureTargetExists(targetType, req.MovieID, req.FranchiseID); err != nil {
		return dto.ExpectationVoteView{}, err
	}

	var saved models.ExpectationVote
	err := s.rep.Transaction(func(tx repository.Repository) error {
		query := tx.Where("user_id = ? AND target_type = ?", userID, targetType)
		if targetType == expectationTargetMovie {
			query = query.Where("film_id = ?", *req.MovieID)
		} else {
			query = query.Where("franchise_id = ?", *req.FranchiseID)
		}

		err := query.First(&saved).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		saved.UserID = userID
		saved.TargetType = targetType
		saved.VoteType = voteType
		saved.Comment = strings.TrimSpace(req.Comment)
		if targetType == expectationTargetMovie {
			saved.MovieID = req.MovieID
			saved.FranchiseID = nil
		} else {
			saved.MovieID = nil
			saved.FranchiseID = req.FranchiseID
		}
		if voteType == expectationVoteScore {
			score := *req.Score
			saved.Score = &score
		} else {
			saved.Score = nil
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&saved).Error
		}
		return tx.Save(&saved).Error
	})
	if err != nil {
		return dto.ExpectationVoteView{}, err
	}

	return s.loadVoteView(saved.ID)
}

func (s *expectationService) Summary(targetType string, targetID uint) (dto.ExpectationSummaryView, error) {
	targetType = strings.TrimSpace(targetType)
	if targetType != expectationTargetMovie && targetType != expectationTargetFranchise {
		return dto.ExpectationSummaryView{}, appErrors.ErrInvalidExpectationTargetType
	}

	if err := s.ensureTargetExists(targetType, pointerTo(targetID, targetType == expectationTargetMovie), pointerTo(targetID, targetType == expectationTargetFranchise)); err != nil {
		return dto.ExpectationSummaryView{}, err
	}

	type aggregate struct {
		Avg          float64
		NumericCount int64
		RefuseCount  int64
	}

	var row aggregate
	query := s.rep.Model(&models.ExpectationVote{}).
		Select(
			"COALESCE(AVG(CASE WHEN vote_type = ? THEN score END), 0) AS avg, "+
				"SUM(CASE WHEN vote_type = ? THEN 1 ELSE 0 END) AS numeric_count, "+
				"SUM(CASE WHEN vote_type = ? THEN 1 ELSE 0 END) AS refuse_count",
			expectationVoteScore,
			expectationVoteScore,
			expectationVoteRefuse,
		).
		Where("target_type = ?", targetType)

	if targetType == expectationTargetMovie {
		query = query.Where("film_id = ?", targetID)
	} else {
		query = query.Where("franchise_id = ?", targetID)
	}

	if err := query.Scan(&row).Error; err != nil {
		return dto.ExpectationSummaryView{}, err
	}

	return dto.ExpectationSummaryView{
		TargetType:      targetType,
		TargetID:        targetID,
		Avg:             row.Avg,
		NumericCount:    row.NumericCount,
		RefuseCount:     row.RefuseCount,
		ThresholdPassed: row.NumericCount > 0 && row.Avg >= 5.0,
	}, nil
}

func (s *expectationService) ListByUser(userID uint) ([]dto.ExpectationVoteView, error) {
	var votes []models.ExpectationVote
	if err := s.rep.Preload("Movie").Preload("Franchise").Where("user_id = ?", userID).Order("updated_at DESC, id DESC").Find(&votes).Error; err != nil {
		return nil, err
	}

	items := make([]dto.ExpectationVoteView, 0, len(votes))
	for _, vote := range votes {
		items = append(items, mapExpectationVote(vote))
	}
	return items, nil
}

func (s *expectationService) ensureTargetExists(targetType string, movieID *uint, franchiseID *uint) error {
	switch targetType {
	case expectationTargetMovie:
		var film models.Film
		if err := s.rep.First(&film, *movieID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrFilmNotFound
			}
			return err
		}
	case expectationTargetFranchise:
		var franchise models.Franchise
		if err := s.rep.First(&franchise, *franchiseID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrFranchiseNotFound
			}
			return err
		}
	default:
		return appErrors.ErrInvalidExpectationTargetType
	}
	return nil
}

func (s *expectationService) loadVoteView(id uint) (dto.ExpectationVoteView, error) {
	var vote models.ExpectationVote
	if err := s.rep.Preload("Movie").Preload("Franchise").First(&vote, id).Error; err != nil {
		return dto.ExpectationVoteView{}, err
	}
	return mapExpectationVote(vote), nil
}

func mapExpectationVote(vote models.ExpectationVote) dto.ExpectationVoteView {
	targetTitle := vote.Franchise.Title
	if vote.TargetType == expectationTargetMovie {
		targetTitle = vote.Movie.Title
	}

	return dto.ExpectationVoteView{
		ID:          vote.ID,
		TargetType:  vote.TargetType,
		MovieID:     vote.MovieID,
		FranchiseID: vote.FranchiseID,
		TargetTitle: targetTitle,
		VoteType:    vote.VoteType,
		Score:       vote.Score,
		Comment:     vote.Comment,
	}
}

func validateExpectationRequest(req dto.UpsertExpectationRequest, targetType, voteType string) error {
	switch targetType {
	case expectationTargetMovie:
		if req.MovieID == nil || *req.MovieID == 0 || req.FranchiseID != nil {
			return appErrors.ErrInvalidExpectationTarget
		}
	case expectationTargetFranchise:
		if req.FranchiseID == nil || *req.FranchiseID == 0 || req.MovieID != nil {
			return appErrors.ErrInvalidExpectationTarget
		}
	default:
		return appErrors.ErrInvalidExpectationTargetType
	}

	switch voteType {
	case expectationVoteScore:
		if req.Score == nil || *req.Score < 1 || *req.Score > 10 {
			return appErrors.ErrInvalidExpectationScore
		}
	case expectationVoteRefuse:
		if req.Score != nil {
			return appErrors.ErrUnexpectedExpectationScore
		}
	default:
		return appErrors.ErrInvalidExpectationVoteType
	}

	return nil
}

func pointerTo(id uint, enabled bool) *uint {
	if !enabled {
		return nil
	}
	value := id
	return &value
}
