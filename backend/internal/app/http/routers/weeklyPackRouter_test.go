package routers

import (
	"cinema/internal/app/http/handlers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeWeeklyPackHandler struct{}

func (fakeWeeklyPackHandler) List(c *gin.Context)         { c.Status(http.StatusOK) }
func (fakeWeeklyPackHandler) GetByID(c *gin.Context)      { c.Status(http.StatusOK) }
func (fakeWeeklyPackHandler) UpsertVote(c *gin.Context)   { c.Status(http.StatusOK) }
func (fakeWeeklyPackHandler) MeVotes(c *gin.Context)      { c.Status(http.StatusOK) }
func (fakeWeeklyPackHandler) Create(c *gin.Context)       { c.Status(http.StatusCreated) }
func (fakeWeeklyPackHandler) AddMovie(c *gin.Context)     { c.Status(http.StatusCreated) }
func (fakeWeeklyPackHandler) UpdateStatus(c *gin.Context) { c.Status(http.StatusOK) }

var _ handlers.WeeklyPackHandler = fakeWeeklyPackHandler{}

func TestRegisterWeeklyPackRoutesPublicEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterWeeklyPackRoutes(r, fakeWeeklyPackHandler{}, "secret")

	req := httptest.NewRequest(http.MethodGet, "/weekly-packs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRegisterWeeklyPackRoutesVoteRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterWeeklyPackRoutes(r, fakeWeeklyPackHandler{}, "secret")

	req := httptest.NewRequest(http.MethodPost, "/weekly-packs/1/votes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
