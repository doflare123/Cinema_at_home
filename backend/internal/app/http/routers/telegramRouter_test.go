package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeTelegramNotificationHandler struct{}

func (fakeTelegramNotificationHandler) List(c *gin.Context)    { c.Status(http.StatusOK) }
func (fakeTelegramNotificationHandler) Enqueue(c *gin.Context) { c.Status(http.StatusCreated) }
func (fakeTelegramNotificationHandler) Retry(c *gin.Context)   { c.Status(http.StatusOK) }

func TestRegisterTelegramNotificationRoutesRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterTelegramNotificationRoutes(r, fakeTelegramNotificationHandler{}, "secret")

	req := httptest.NewRequest(http.MethodGet, "/admin/telegram/notifications", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRegisterTelegramNotificationRoutesRejectMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterTelegramNotificationRoutes(r, fakeTelegramNotificationHandler{}, "secret")

	req := httptest.NewRequest(http.MethodPost, "/admin/telegram/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "secret", 1, 1, "active", "access"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRegisterTelegramNotificationRoutesAllowAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterTelegramNotificationRoutes(r, fakeTelegramNotificationHandler{}, "secret")

	req := httptest.NewRequest(http.MethodPost, "/admin/telegram/notifications/10/retry", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "secret", 1, 2, "active", "access"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
