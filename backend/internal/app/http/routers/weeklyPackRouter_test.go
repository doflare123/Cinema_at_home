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
func (fakeWeeklyPackHandler) Current(c *gin.Context)      { c.Status(http.StatusOK) }
func (fakeWeeklyPackHandler) GetByID(c *gin.Context)      { c.Status(http.StatusOK) }
func (fakeWeeklyPackHandler) UpsertVote(c *gin.Context)   { c.Status(http.StatusOK) }
func (fakeWeeklyPackHandler) MeVotes(c *gin.Context)      { c.Status(http.StatusOK) }
func (fakeWeeklyPackHandler) MeVoteLimits(c *gin.Context) { c.Status(http.StatusOK) }
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

func TestRegisterWeeklyPackRoutesCurrentEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterWeeklyPackRoutes(r, fakeWeeklyPackHandler{}, "secret")

	req := httptest.NewRequest(http.MethodGet, "/weekly-packs/current", nil)
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

func TestRegisterWeeklyPackRoutesVoteRejectsWrongRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterWeeklyPackRoutes(r, fakeWeeklyPackHandler{}, "secret")

	req := httptest.NewRequest(http.MethodPost, "/weekly-packs/1/votes", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "secret", 1, 3, "active", "access"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRegisterWeeklyPackRoutesVoteAllowsMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterWeeklyPackRoutes(r, fakeWeeklyPackHandler{}, "secret")

	req := httptest.NewRequest(http.MethodPost, "/weekly-packs/1/votes", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "secret", 1, 1, "active", "access"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRegisterWeeklyPackRoutesVoteLimitsRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterWeeklyPackRoutes(r, fakeWeeklyPackHandler{}, "secret")

	req := httptest.NewRequest(http.MethodGet, "/weekly-packs/1/votes/me/limits", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
