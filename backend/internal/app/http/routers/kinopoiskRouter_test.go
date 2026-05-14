package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeKinopoiskHandler struct{}

func (fakeKinopoiskHandler) Search(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"route": "kinopoisk_search"})
}

func TestRegisterKinopoiskRoutesRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterKinopoiskRoutes(r, fakeKinopoiskHandler{}, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/kinopoisk/search?query=matrix", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
