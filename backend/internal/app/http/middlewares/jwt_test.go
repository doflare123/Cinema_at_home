package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestGetUintClaimFromFloat64(t *testing.T) {
	claims := jwt.MapClaims{"user_id": float64(42)}
	v, ok := getUintClaim(claims, "user_id")
	if !ok {
		t.Fatal("expected claim to be parsed")
	}
	if v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
}

func TestJWTAuthMiddlewareSetsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"
	claims := jwt.MapClaims{
		"user_id":  float64(1),
		"role_id":  float64(2),
		"username": "tester",
		"status":   "active",
		"type":     "access",
		"exp":      time.Now().Add(time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	r := gin.New()
	r.Use(JWTAuthMiddleware(secret))
	r.GET("/private", func(c *gin.Context) {
		if c.GetUint("user_id") != 1 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "wrong user_id"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequireActiveStatusBlocksPending(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("status", "pending")
		c.Next()
	})
	r.Use(RequireActiveStatus())
	r.GET("/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestJWTAuthMiddlewareRejectsRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"
	claims := jwt.MapClaims{
		"user_id": float64(1),
		"role_id": float64(2),
		"status":  "active",
		"type":    "refresh",
		"exp":     time.Now().Add(time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	r := gin.New()
	r.Use(JWTAuthMiddleware(secret))
	r.GET("/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
