package services

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"cinema/internal/repository"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	reviewModeSimple   = "simple"
	reviewModeCriteria = "criteria"
)

type ReviewService interface {
	ListByFilm(filmID uint) ([]dto.ReviewView, error)
	GetByFilmAndUser(filmID, userID uint) (dto.ReviewView, error)
	Upsert(filmID, userID uint, req dto.UpsertReviewRequest) (dto.ReviewView, error)
}

type reviewService struct {
	rep repository.Repository
}

func NewReviewService(rep repository.Repository) ReviewService {
	return &reviewService{rep: rep}
}

func (s *reviewService) ListByFilm(filmID uint) ([]dto.ReviewView, error) {
	if err := s.ensureFilmExists(filmID); err != nil {
		return nil, err
	}

	var reviews []models.Review
	if err := s.rep.Preload("User").Where("film_id = ?", filmID).Order("updated_at DESC, id DESC").Find(&reviews).Error; err != nil {
		return nil, err
	}

	out := make([]dto.ReviewView, 0, len(reviews))
	for _, review := range reviews {
		view, err := mapReview(review)
		if err != nil {
			return nil, err
		}
		out = append(out, view)
	}
	return out, nil
}

func (s *reviewService) GetByFilmAndUser(filmID, userID uint) (dto.ReviewView, error) {
	if err := s.ensureFilmExists(filmID); err != nil {
		return dto.ReviewView{}, err
	}

	var review models.Review
	if err := s.rep.Preload("User").Where("film_id = ? AND user_id = ?", filmID, userID).First(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ReviewView{}, appErrors.ErrReviewNotFound
		}
		return dto.ReviewView{}, err
	}

	return mapReview(review)
}

func (s *reviewService) Upsert(filmID, userID uint, req dto.UpsertReviewRequest) (dto.ReviewView, error) {
	if err := s.ensureFilmExists(filmID); err != nil {
		return dto.ReviewView{}, err
	}

	payload, err := buildReviewPayload(req)
	if err != nil {
		return dto.ReviewView{}, err
	}

	err = s.rep.Transaction(func(tx repository.Repository) error {
		now := time.Now()
		review := models.Review{
			FilmID:             filmID,
			UserID:             userID,
			Mode:               payload.Mode,
			Score:              payload.Score,
			FinalScore:         payload.FinalScore,
			CriteriaScoresJSON: payload.CriteriaScoresJSON,
			Comment:            payload.Comment,
		}
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "film_id"},
				{Name: "user_id"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"mode":            review.Mode,
				"score":           review.Score,
				"final_score":     review.FinalScore,
				"criteria_scores": review.CriteriaScoresJSON,
				"comment":         review.Comment,
				"updated_at":      now,
			}),
		}).Create(&review).Error
	})
	if err != nil {
		return dto.ReviewView{}, err
	}

	return s.GetByFilmAndUser(filmID, userID)
}

func (s *reviewService) ensureFilmExists(filmID uint) error {
	var film models.Film
	if err := s.rep.First(&film, filmID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrFilmNotFound
		}
		return err
	}
	return nil
}

type reviewPayload struct {
	Mode               string
	Score              *int
	FinalScore         float64
	CriteriaScoresJSON string
	Comment            string
}

func buildReviewPayload(req dto.UpsertReviewRequest) (reviewPayload, error) {
	if req.FinalScore != nil {
		return reviewPayload{}, appErrors.ErrManualReviewFinalScoreEdit
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	comment := strings.TrimSpace(req.Comment)

	switch mode {
	case reviewModeSimple:
		if req.Score == nil {
			return reviewPayload{}, appErrors.ErrReviewScoreRequired
		}
		if *req.Score < 1 || *req.Score > 10 {
			return reviewPayload{}, appErrors.ErrInvalidReviewScore
		}
		if len(req.CriteriaScores) > 0 {
			return reviewPayload{}, appErrors.ErrUnexpectedReviewCriteria
		}
		score := *req.Score
		return reviewPayload{
			Mode:               mode,
			Score:              &score,
			FinalScore:         float64(score),
			CriteriaScoresJSON: "{}",
			Comment:            comment,
		}, nil
	case reviewModeCriteria:
		if req.Score != nil {
			return reviewPayload{}, appErrors.ErrUnexpectedReviewScore
		}
		normalized, avg, err := normalizeCriteriaScores(req.CriteriaScores)
		if err != nil {
			return reviewPayload{}, err
		}
		raw, err := json.Marshal(normalized)
		if err != nil {
			return reviewPayload{}, err
		}
		return reviewPayload{
			Mode:               mode,
			FinalScore:         avg,
			CriteriaScoresJSON: string(raw),
			Comment:            comment,
		}, nil
	default:
		return reviewPayload{}, appErrors.ErrInvalidReviewMode
	}
}

func normalizeCriteriaScores(criteria map[string]int) (map[string]int, float64, error) {
	if len(criteria) == 0 {
		return nil, 0, appErrors.ErrReviewCriteriaRequired
	}

	keys := make([]string, 0, len(criteria))
	for key := range criteria {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	normalized := make(map[string]int, len(criteria))
	total := 0
	for _, key := range keys {
		trimmed := strings.TrimSpace(key)
		score := criteria[key]
		if trimmed == "" || score < 1 || score > 10 {
			return nil, 0, appErrors.ErrInvalidReviewCriterion
		}
		if _, exists := normalized[trimmed]; exists {
			return nil, 0, appErrors.ErrInvalidReviewCriterion
		}
		normalized[trimmed] = score
		total += score
	}

	return normalized, roundReviewScore(float64(total) / float64(len(normalized))), nil
}

func mapReview(review models.Review) (dto.ReviewView, error) {
	criteria, err := decodeCriteriaScores(review.CriteriaScoresJSON)
	if err != nil {
		return dto.ReviewView{}, err
	}

	return dto.ReviewView{
		ID:             review.ID,
		FilmID:         review.FilmID,
		UserID:         review.UserID,
		Username:       review.User.Username,
		DisplayName:    review.User.DisplayName,
		Mode:           review.Mode,
		Score:          review.Score,
		CriteriaScores: criteria,
		FinalScore:     review.FinalScore,
		Comment:        review.Comment,
		CreatedAt:      review.CreatedAt,
		UpdatedAt:      review.UpdatedAt,
	}, nil
}

func decodeCriteriaScores(raw string) (map[string]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return nil, nil
	}

	var out map[string]int
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

func roundReviewScore(score float64) float64 {
	return math.Round(score*100) / 100
}
