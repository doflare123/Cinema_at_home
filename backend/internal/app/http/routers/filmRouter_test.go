package routers

import (
	"cinema/internal/repository"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeFilmHandler struct{}

func (fakeFilmHandler) List(c *gin.Context)    { c.JSON(http.StatusOK, gin.H{"route": "list_movies"}) }
func (fakeFilmHandler) GetByID(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"route": "get_movie"}) }
func (fakeFilmHandler) CreateFilm(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"route": "create_film"})
}

func TestRegisterFilmRoutesMoviesListIsPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterFilmRoutes(r, fakeFilmHandler{}, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/movies", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRegisterFilmRoutesMovieDetailIsPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterFilmRoutes(r, fakeFilmHandler{}, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/movies/42", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRegisterFilmRoutesCreateStillRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterFilmRoutes(r, fakeFilmHandler{}, "test-secret", repository.Repository(nil))

	req := httptest.NewRequest(http.MethodPost, "/film/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
