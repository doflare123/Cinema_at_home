package middlewares

import (
	"cinema/internal/models"
	"cinema/internal/repository"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func TestJWTAuthMiddlewareRejectsUserBlockedAfterTokenIssued(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rep, user := newMiddlewareTestRepository(t)
	secret := "test-secret"
	claims := jwt.MapClaims{
		"user_id":  float64(user.ID),
		"role_id":  float64(user.RoleID),
		"username": user.Username,
		"status":   "active",
		"type":     "access",
		"exp":      time.Now().Add(time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	if err := rep.Model(&models.User{}).Where("id = ?", user.ID).Update("status", "blocked").Error; err != nil {
		t.Fatalf("failed to block user: %v", err)
	}

	r := gin.New()
	r.Use(JWTAuthMiddleware(secret, rep), RequireActiveStatus())
	r.GET("/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestJWTAuthMiddlewareUsesActualRoleFromDB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rep, user := newMiddlewareTestRepository(t)
	secret := "test-secret"
	claims := jwt.MapClaims{
		"user_id":  float64(user.ID),
		"role_id":  float64(2),
		"username": user.Username,
		"status":   "active",
		"type":     "access",
		"exp":      time.Now().Add(time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// DB role stays member(1), token claims admin(2). Middleware must use DB role and deny.
	r := gin.New()
	r.Use(JWTAuthMiddleware(secret, rep), RequireActiveStatus(), RequireRoles(2))
	r.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func newMiddlewareTestRepository(t *testing.T) (repository.Repository, *models.User) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%d?mode=memory&cache=shared", time.Now().UnixNano())), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	rep := &middlewareTestRepository{db: db}
	for _, model := range []interface{}{&models.Role{}, &models.User{}} {
		if err := rep.AutoMigrate(model); err != nil {
			t.Fatalf("failed to migrate test db: %v", err)
		}
	}

	user := &models.User{
		Username:    "tester",
		Password:    "unused",
		DisplayName: "Tester",
		RoleID:      1,
		Status:      "active",
	}
	if err := rep.Create(user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	return rep, user
}

type middlewareTestRepository struct {
	db *gorm.DB
}

func (r *middlewareTestRepository) Model(value interface{}) *gorm.DB { return r.db.Model(value) }
func (r *middlewareTestRepository) Select(query interface{}, args ...interface{}) *gorm.DB {
	return r.db.Select(query, args...)
}
func (r *middlewareTestRepository) Find(out interface{}, where ...interface{}) *gorm.DB {
	return r.db.Find(out, where...)
}
func (r *middlewareTestRepository) Exec(query string, values ...interface{}) *gorm.DB {
	return r.db.Exec(query, values...)
}
func (r *middlewareTestRepository) First(out interface{}, where ...interface{}) *gorm.DB {
	return r.db.First(out, where...)
}
func (r *middlewareTestRepository) Raw(query string, values ...interface{}) *gorm.DB {
	return r.db.Raw(query, values...)
}
func (r *middlewareTestRepository) Create(value interface{}) *gorm.DB  { return r.db.Create(value) }
func (r *middlewareTestRepository) Save(value interface{}) *gorm.DB    { return r.db.Save(value) }
func (r *middlewareTestRepository) Updates(value interface{}) *gorm.DB { return r.db.Updates(value) }
func (r *middlewareTestRepository) Delete(value interface{}) *gorm.DB  { return r.db.Delete(value) }
func (r *middlewareTestRepository) Where(query interface{}, args ...interface{}) *gorm.DB {
	return r.db.Where(query, args...)
}
func (r *middlewareTestRepository) Preload(column string, conditions ...interface{}) *gorm.DB {
	return r.db.Preload(column, conditions...)
}
func (r *middlewareTestRepository) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB {
	return r.db.Scopes(funcs...)
}
func (r *middlewareTestRepository) ScanRows(rows *sql.Rows, result interface{}) error {
	return r.db.ScanRows(rows, result)
}
func (r *middlewareTestRepository) Clauses(conds ...clause.Expression) *gorm.DB {
	return r.db.Clauses(conds...)
}
func (r *middlewareTestRepository) AutoMigrate(value interface{}) error {
	return r.db.AutoMigrate(value)
}
func (r *middlewareTestRepository) DropTableIfExists(value interface{}) error {
	return r.db.Migrator().DropTable(value)
}
func (r *middlewareTestRepository) GetSQLDB() (*sql.DB, error) { return r.db.DB() }
func (r *middlewareTestRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
func (r *middlewareTestRepository) Transaction(fc func(tx repository.Repository) error) (err error) {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fc(&middlewareTestRepository{db: tx})
	})
}
