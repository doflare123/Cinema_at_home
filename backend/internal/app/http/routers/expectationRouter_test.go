package routers

import (
	"cinema/internal/repository"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeExpectationHandler struct{}

func (fakeExpectationHandler) Upsert(c *gin.Context)  { c.Status(http.StatusOK) }
func (fakeExpectationHandler) Summary(c *gin.Context) { c.Status(http.StatusOK) }
func (fakeExpectationHandler) Me(c *gin.Context)      { c.Status(http.StatusOK) }

func TestRegisterExpectationRoutesWriteRejectsWrongRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterExpectationRoutes(r, fakeExpectationHandler{}, "secret", repository.Repository(nil))

	req := httptest.NewRequest(http.MethodPost, "/expectations", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "secret", 1, 3, "active", "access"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRegisterExpectationRoutesWriteAllowsMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterExpectationRoutes(r, fakeExpectationHandler{}, "secret", repository.Repository(nil))

	req := httptest.NewRequest(http.MethodPost, "/expectations", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "secret", 1, 1, "active", "access"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
