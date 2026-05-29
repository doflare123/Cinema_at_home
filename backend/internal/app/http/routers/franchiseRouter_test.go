package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeFranchiseHandler struct{}

func (fakeFranchiseHandler) List(c *gin.Context)    { c.Status(http.StatusOK) }
func (fakeFranchiseHandler) GetByID(c *gin.Context) { c.Status(http.StatusOK) }
func (fakeFranchiseHandler) Create(c *gin.Context)  { c.Status(http.StatusCreated) }
func (fakeFranchiseHandler) AddMovie(c *gin.Context) {
	c.Status(http.StatusCreated)
}

func TestRegisterFranchiseRoutesAdminRejectsMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterFranchiseRoutes(r, fakeFranchiseHandler{}, "secret")

	req := httptest.NewRequest(http.MethodPost, "/admin/franchises", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "secret", 1, 1, "active", "access"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRegisterFranchiseRoutesAdminAllowsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterFranchiseRoutes(r, fakeFranchiseHandler{}, "secret")

	req := httptest.NewRequest(http.MethodPost, "/admin/franchises", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "secret", 1, 2, "active", "access"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
