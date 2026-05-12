package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type fakeAuthHandler struct {
	secret string
}

func (h fakeAuthHandler) Register(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"route": "register"}) }
func (h fakeAuthHandler) Login(c *gin.Context)    { c.JSON(http.StatusOK, gin.H{"route": "login"}) }
func (h fakeAuthHandler) Telegram(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"route": "telegram"}) }
func (h fakeAuthHandler) Refresh(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"route": "refresh"}) }
func (h fakeAuthHandler) Me(c *gin.Context)       { c.JSON(http.StatusOK, gin.H{"route": "me"}) }
func (h fakeAuthHandler) ListUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"route": "list_users"})
}
func (h fakeAuthHandler) UpdateUserStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"route": "update_user_status"})
}
func (h fakeAuthHandler) JWTSecret() string { return h.secret }

func TestRegisterAuthRoutesMeRequiresActiveAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterAuthRoutes(r, fakeAuthHandler{secret: "test-secret"})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "test-secret", 1, "active", "access"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRegisterAuthRoutesMeRejectsRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterAuthRoutes(r, fakeAuthHandler{secret: "test-secret"})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "test-secret", 1, "active", "refresh"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRegisterAuthRoutesAdminUsersRequiresAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterAuthRoutes(r, fakeAuthHandler{secret: "test-secret"})

	req := httptest.NewRequest(http.MethodGet, "/admin/users?status=pending", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "test-secret", 1, 1, "active", "access"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRegisterAuthRoutesAdminUsersAllowsAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterAuthRoutes(r, fakeAuthHandler{secret: "test-secret"})

	req := httptest.NewRequest(http.MethodGet, "/admin/users?status=pending", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "test-secret", 1, 2, "active", "access"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func signTestToken(t *testing.T, secret string, userID uint, status, tokenType string) string {
	t.Helper()
	return signTestTokenWithRole(t, secret, userID, 1, status, tokenType)
}

func signTestTokenWithRole(t *testing.T, secret string, userID, roleID uint, status, tokenType string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  float64(userID),
		"role_id":  float64(roleID),
		"username": "tester",
		"status":   status,
		"type":     tokenType,
		"exp":      time.Now().Add(time.Minute).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenStr
}
