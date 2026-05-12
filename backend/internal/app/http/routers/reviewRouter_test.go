package routers

import (
	"cinema/internal/repository"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeReviewHandler struct{}

func (fakeReviewHandler) ListByFilm(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"route": "list_reviews"})
}
func (fakeReviewHandler) Me(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"route": "my_review"}) }
func (fakeReviewHandler) Upsert(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"route": "upsert_review"})
}

func TestRegisterReviewRoutesListIsPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterReviewRoutes(r, fakeReviewHandler{}, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/reviews/films/42", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRegisterReviewRoutesMeRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterReviewRoutes(r, fakeReviewHandler{}, "test-secret", repository.Repository(nil))

	req := httptest.NewRequest(http.MethodGet, "/reviews/films/42/me", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
