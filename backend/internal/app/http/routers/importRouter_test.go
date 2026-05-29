package routers

import (
	"cinema/internal/app/http/handlers"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeImportHandler struct{}

func (fakeImportHandler) ImportCatalogExcel(c *gin.Context) { c.Status(http.StatusOK) }

var _ handlers.ImportHandler = fakeImportHandler{}

func TestRegisterImportRoutesRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterImportRoutes(r, fakeImportHandler{}, "secret")

	req := httptest.NewRequest(http.MethodPost, "/admin/import/excel/catalog", strings.NewReader(""))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRegisterImportRoutesForbidsMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterImportRoutes(r, fakeImportHandler{}, "secret")

	req := httptest.NewRequest(http.MethodPost, "/admin/import/excel/catalog", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "secret", 1, 1, "active", "access"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRegisterImportRoutesAllowsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterImportRoutes(r, fakeImportHandler{}, "secret")

	req := httptest.NewRequest(http.MethodPost, "/admin/import/excel/catalog", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "secret", 1, 2, "active", "access"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
