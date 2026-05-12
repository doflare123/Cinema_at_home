package handlers

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models/dto"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubFilmService struct {
	list    func() ([]dto.FilmView, error)
	getByID func(id uint) (dto.FilmView, error)
}

func (s stubFilmService) Create(title string) error {
	return nil
}

func (s stubFilmService) List() ([]dto.FilmView, error) {
	if s.list != nil {
		return s.list()
	}
	return nil, nil
}

func (s stubFilmService) GetByID(id uint) (dto.FilmView, error) {
	if s.getByID != nil {
		return s.getByID(id)
	}
	return dto.FilmView{}, nil
}

func TestFilmHandlerListReturnsMoviesPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewFilmHandler(stubFilmService{
		list: func() ([]dto.FilmView, error) {
			return []dto.FilmView{{
				ID:          1,
				Title:       "Movie",
				Description: "Description",
				Poster:      "poster.jpg",
				ReleaseDate: 2024,
			}}, nil
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/movies", nil)

	handler.List(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response struct {
		Movies []dto.FilmView `json:"movies"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(response.Movies) != 1 {
		t.Fatalf("expected 1 movie, got %d", len(response.Movies))
	}
	if response.Movies[0].Description != "Description" || response.Movies[0].Poster != "poster.jpg" || response.Movies[0].ReleaseDate != 2024 {
		t.Fatalf("unexpected movie payload: %+v", response.Movies[0])
	}
}

func TestFilmHandlerGetByIDReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewFilmHandler(stubFilmService{
		getByID: func(id uint) (dto.FilmView, error) {
			return dto.FilmView{}, appErrors.ErrFilmNotFound
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: "999"}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/movies/999", nil)

	handler.GetByID(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}
