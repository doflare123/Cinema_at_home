package handlers

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models/dto"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type stubReviewService struct {
	listByFilm       func(filmID uint) ([]dto.ReviewView, error)
	getByFilmAndUser func(filmID, userID uint) (dto.ReviewView, error)
	upsert           func(filmID, userID uint, req dto.UpsertReviewRequest) (dto.ReviewView, error)
}

func (s stubReviewService) ListByFilm(filmID uint) ([]dto.ReviewView, error) {
	if s.listByFilm != nil {
		return s.listByFilm(filmID)
	}
	return nil, nil
}

func (s stubReviewService) GetByFilmAndUser(filmID, userID uint) (dto.ReviewView, error) {
	if s.getByFilmAndUser != nil {
		return s.getByFilmAndUser(filmID, userID)
	}
	return dto.ReviewView{}, nil
}

func (s stubReviewService) Upsert(filmID, userID uint, req dto.UpsertReviewRequest) (dto.ReviewView, error) {
	if s.upsert != nil {
		return s.upsert(filmID, userID, req)
	}
	return dto.ReviewView{}, nil
}

func TestReviewHandlerListByFilmReturnsPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewReviewHandler(stubReviewService{
		listByFilm: func(filmID uint) ([]dto.ReviewView, error) {
			return []dto.ReviewView{{
				ID:          1,
				FilmID:      filmID,
				UserID:      3,
				DisplayName: "Reviewer",
				Mode:        "simple",
				FinalScore:  9,
				Comment:     "great",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}}, nil
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "filmId", Value: "42"}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/reviews/films/42", nil)

	handler.ListByFilm(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response struct {
		Reviews []dto.ReviewView `json:"reviews"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(response.Reviews) != 1 || response.Reviews[0].FinalScore != 9 {
		t.Fatalf("unexpected response payload: %+v", response)
	}
}

func TestReviewHandlerMeReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewReviewHandler(stubReviewService{
		getByFilmAndUser: func(filmID, userID uint) (dto.ReviewView, error) {
			return dto.ReviewView{}, appErrors.ErrReviewNotFound
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "filmId", Value: "42"}}
	ctx.Set("user_id", uint(7))
	ctx.Request = httptest.NewRequest(http.MethodGet, "/reviews/films/42/me", nil)

	handler.Me(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestReviewHandlerUpsertMapsValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewReviewHandler(stubReviewService{
		upsert: func(filmID, userID uint, req dto.UpsertReviewRequest) (dto.ReviewView, error) {
			return dto.ReviewView{}, appErrors.ErrManualReviewFinalScoreEdit
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "filmId", Value: "42"}}
	ctx.Set("user_id", uint(7))
	ctx.Request = httptest.NewRequest(http.MethodPost, "/reviews/films/42", strings.NewReader(`{"mode":"criteria","criteria_scores":{"story":8}}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Upsert(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}
