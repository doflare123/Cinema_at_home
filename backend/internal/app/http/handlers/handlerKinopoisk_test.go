package handlers

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models/dto"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubKinopoiskService struct {
	search func(query string, limit int) ([]dto.KinopoiskSearchResult, error)
}

func (s stubKinopoiskService) Search(query string, limit int) ([]dto.KinopoiskSearchResult, error) {
	if s.search != nil {
		return s.search(query, limit)
	}
	return nil, nil
}

func TestKinopoiskHandlerSearchReturnsPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewKinopoiskHandler(stubKinopoiskService{
		search: func(query string, limit int) ([]dto.KinopoiskSearchResult, error) {
			if query != "matrix" || limit != 2 {
				t.Fatalf("unexpected query/limit: %q %d", query, limit)
			}
			return []dto.KinopoiskSearchResult{{ProviderMovieID: "1", Title: "Matrix", Source: "kinopoisk"}}, nil
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/kinopoisk/search?q=matrix&limit=2", nil)

	handler.Search(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response struct {
		Results []dto.KinopoiskSearchResult `json:"results"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Results) != 1 || response.Results[0].ProviderMovieID != "1" || response.Results[0].Source != "kinopoisk" {
		t.Fatalf("unexpected response contract: %+v", response)
	}
}

func TestKinopoiskHandlerSearchMapsProviderError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewKinopoiskHandler(stubKinopoiskService{
		search: func(query string, limit int) ([]dto.KinopoiskSearchResult, error) {
			return nil, fmt.Errorf("%w: upstream secret detail", appErrors.ErrKinopoiskSearchFailed)
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/kinopoisk/search?query=matrix", nil)

	handler.Search(ctx)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), appErrors.ErrKinopoiskSearchFailed.Error()) {
		t.Fatalf("expected generic kinopoisk error, got %s", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "upstream secret detail") {
		t.Fatalf("response leaked upstream error detail: %s", recorder.Body.String())
	}
}
