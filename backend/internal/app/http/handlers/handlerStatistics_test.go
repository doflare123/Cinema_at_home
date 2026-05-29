package handlers

import (
	"cinema/internal/models/dto"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubStatisticsService struct {
	summary func() (dto.StatisticsSummaryView, error)
}

func (s stubStatisticsService) Summary() (dto.StatisticsSummaryView, error) {
	if s.summary != nil {
		return s.summary()
	}
	return dto.StatisticsSummaryView{}, nil
}

func TestStatisticsHandlerSummaryReturnsPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStatisticsHandler(stubStatisticsService{
		summary: func() (dto.StatisticsSummaryView, error) {
			return dto.StatisticsSummaryView{MoviesTotal: 5}, nil
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/statistics/summary", nil)

	handler.Summary(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestStatisticsHandlerSummaryReturnsInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStatisticsHandler(stubStatisticsService{
		summary: func() (dto.StatisticsSummaryView, error) {
			return dto.StatisticsSummaryView{}, errors.New("db failed")
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/statistics/summary", nil)

	handler.Summary(ctx)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
