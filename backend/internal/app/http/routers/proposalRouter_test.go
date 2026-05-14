package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeMovieProposalHandler struct{}

func (fakeMovieProposalHandler) Create(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"route": "create_proposal"})
}
func (fakeMovieProposalHandler) My(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"route": "my_proposals"})
}
func (fakeMovieProposalHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"route": "list_proposals"})
}
func (fakeMovieProposalHandler) Moderate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"route": "moderate_proposal"})
}

func TestRegisterMovieProposalRoutesRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterMovieProposalRoutes(r, fakeMovieProposalHandler{}, "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/proposals", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRegisterMovieProposalRoutesAdminListRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterMovieProposalRoutes(r, fakeMovieProposalHandler{}, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/admin/proposals", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
