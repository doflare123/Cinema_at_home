package routers

import (
	"cinema/internal/models"
	"cinema/internal/repository"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type fakeKinopoiskHandler struct{}

var kinopoiskRouterTestDBCounter uint64

func (fakeKinopoiskHandler) Search(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"route": "kinopoisk_search"})
}

func TestRegisterKinopoiskRoutesRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterKinopoiskRoutes(r, fakeKinopoiskHandler{}, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/kinopoisk/search?query=matrix", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRegisterKinopoiskRoutesRejectInactiveUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterKinopoiskRoutes(r, fakeKinopoiskHandler{}, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/kinopoisk/search?query=matrix", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "test-secret", 1, 1, "pending", "access"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRegisterKinopoiskRoutesRejectWrongRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterKinopoiskRoutes(r, fakeKinopoiskHandler{}, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/kinopoisk/search?query=matrix", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "test-secret", 1, 3, "active", "access"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRegisterKinopoiskRoutesAllowMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterKinopoiskRoutes(r, fakeKinopoiskHandler{}, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/kinopoisk/search?query=matrix", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "test-secret", 1, 1, "active", "access"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRegisterKinopoiskRoutesRepositoryOverridesClaimsToForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rep, userID := newKinopoiskRouterRepository(t, "pending", "member")

	r := gin.New()
	RegisterKinopoiskRoutes(r, fakeKinopoiskHandler{}, "test-secret", rep)

	req := httptest.NewRequest(http.MethodGet, "/kinopoisk/search?query=matrix", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "test-secret", userID, 2, "active", "access"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRegisterKinopoiskRoutesRepositoryAllowsActiveMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rep, userID := newKinopoiskRouterRepository(t, "active", "member")

	r := gin.New()
	RegisterKinopoiskRoutes(r, fakeKinopoiskHandler{}, "test-secret", rep)

	req := httptest.NewRequest(http.MethodGet, "/kinopoisk/search?query=matrix", nil)
	req.Header.Set("Authorization", "Bearer "+signTestTokenWithRole(t, "test-secret", userID, 3, "pending", "access"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

type kinopoiskRouterRepository struct {
	db *gorm.DB
}

func newKinopoiskRouterRepository(t *testing.T, userStatus, roleName string) (repository.Repository, uint) {
	t.Helper()

	db := openKinopoiskRouterSQLiteDB(t)
	rep := &kinopoiskRouterRepository{db: db}

	if err := rep.AutoMigrate(&models.Role{}); err != nil {
		t.Fatalf("migrate roles failed: %v", err)
	}
	if err := rep.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate users failed: %v", err)
	}

	roles := []models.Role{
		{ID: 1, Name: "member"},
		{ID: 2, Name: "admin"},
		{ID: 3, Name: "guest"},
	}
	for _, role := range roles {
		if err := rep.Create(&role).Error; err != nil {
			t.Fatalf("seed role %s failed: %v", role.Name, err)
		}
	}

	roleID := uint(1)
	if roleName == "admin" {
		roleID = 2
	}
	if roleName == "guest" {
		roleID = 3
	}

	user := models.User{
		Username:    "router_user",
		Password:    "x",
		DisplayName: "Router User",
		RoleID:      roleID,
		Status:      userStatus,
	}
	if err := rep.Create(&user).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	return rep, user.ID
}

func openKinopoiskRouterSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()

	id := atomic.AddUint64(&kinopoiskRouterTestDBCounter, 1)
	name := fmt.Sprintf("router-test-%d-%d", time.Now().UnixNano(), id)
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", name)), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	return db
}

func (r *kinopoiskRouterRepository) Model(value interface{}) *gorm.DB { return r.db.Model(value) }
func (r *kinopoiskRouterRepository) Select(query interface{}, args ...interface{}) *gorm.DB {
	return r.db.Select(query, args...)
}
func (r *kinopoiskRouterRepository) Find(out interface{}, where ...interface{}) *gorm.DB {
	return r.db.Find(out, where...)
}
func (r *kinopoiskRouterRepository) Exec(query string, values ...interface{}) *gorm.DB {
	return r.db.Exec(query, values...)
}
func (r *kinopoiskRouterRepository) First(out interface{}, where ...interface{}) *gorm.DB {
	return r.db.First(out, where...)
}
func (r *kinopoiskRouterRepository) Raw(query string, values ...interface{}) *gorm.DB {
	return r.db.Raw(query, values...)
}
func (r *kinopoiskRouterRepository) Create(value interface{}) *gorm.DB  { return r.db.Create(value) }
func (r *kinopoiskRouterRepository) Save(value interface{}) *gorm.DB    { return r.db.Save(value) }
func (r *kinopoiskRouterRepository) Updates(value interface{}) *gorm.DB { return r.db.Updates(value) }
func (r *kinopoiskRouterRepository) Delete(value interface{}) *gorm.DB  { return r.db.Delete(value) }
func (r *kinopoiskRouterRepository) Where(query interface{}, args ...interface{}) *gorm.DB {
	return r.db.Where(query, args...)
}
func (r *kinopoiskRouterRepository) Preload(column string, conditions ...interface{}) *gorm.DB {
	return r.db.Preload(column, conditions...)
}
func (r *kinopoiskRouterRepository) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB {
	return r.db.Scopes(funcs...)
}
func (r *kinopoiskRouterRepository) ScanRows(rows *sql.Rows, result interface{}) error {
	return r.db.ScanRows(rows, result)
}
func (r *kinopoiskRouterRepository) Clauses(conds ...clause.Expression) *gorm.DB {
	return r.db.Clauses(conds...)
}
func (r *kinopoiskRouterRepository) AutoMigrate(value interface{}) error {
	return r.db.AutoMigrate(value)
}
func (r *kinopoiskRouterRepository) DropTableIfExists(value interface{}) error {
	return r.db.Migrator().DropTable(value)
}
func (r *kinopoiskRouterRepository) GetSQLDB() (*sql.DB, error) { return r.db.DB() }
func (r *kinopoiskRouterRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
func (r *kinopoiskRouterRepository) Transaction(fc func(tx repository.Repository) error) (err error) {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fc(&kinopoiskRouterRepository{db: tx})
	})
}
