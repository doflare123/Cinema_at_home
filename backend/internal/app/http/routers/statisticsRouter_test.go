package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeStatisticsHandler struct{}

func (fakeStatisticsHandler) Summary(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"route": "statistics_summary"})
}

func TestRegisterStatisticsRoutesSummaryIsPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterStatisticsRoutes(r, fakeStatisticsHandler{})

	req := httptest.NewRequest(http.MethodGet, "/statistics/summary", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
