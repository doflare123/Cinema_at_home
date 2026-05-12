package services

import (
	"cinema/config"
	"cinema/internal/app/utils"
	"cinema/internal/container"
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"cinema/internal/repository"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var testDBCounter uint64

func TestVerifyTelegramAuthPayloadValid(t *testing.T) {
	botToken := "bot-token"
	req := dto.TelegramAuthRequest{
		TelegramID:       12345,
		TelegramUsername: "tester",
		DisplayName:      "Tester",
		AvatarURL:        "https://example.com/a.png",
		AuthDate:         time.Now().Unix(),
	}
	req.Hash = buildCustomTelegramHash(req, botToken)

	if err := verifyTelegramAuthPayload(req, botToken); err != nil {
		t.Fatalf("expected payload to be valid, got %v", err)
	}
}

func TestVerifyAndExtractTelegramInitDataValid(t *testing.T) {
	botToken := "bot-token"
	vals := url.Values{}
	vals.Set("query_id", "AAEAAAE")
	vals.Set("user", "{\"id\":42,\"first_name\":\"John\",\"last_name\":\"Doe\",\"username\":\"tester\",\"photo_url\":\"https://example.com/avatar.png\"}")
	vals.Set("auth_date", fmt.Sprintf("%d", time.Now().Unix()))
	vals.Set("hash", buildInitDataHash(vals, botToken))

	profile, err := verifyAndExtractTelegramInitData(vals.Encode(), botToken)
	if err != nil {
		t.Fatalf("expected valid init data, got %v", err)
	}
	if profile.TelegramID != 42 {
		t.Fatalf("expected id 42, got %d", profile.TelegramID)
	}
	if profile.DisplayName != "John Doe" {
		t.Fatalf("unexpected display name: %s", profile.DisplayName)
	}
}

func TestVerifyTelegramAuthPayloadInvalidHash(t *testing.T) {
	botToken := "bot-token"
	req := dto.TelegramAuthRequest{
		TelegramID:       12345,
		TelegramUsername: "tester",
		DisplayName:      "Tester",
		AuthDate:         time.Now().Unix(),
		Hash:             "deadbeef",
	}

	if err := verifyTelegramAuthPayload(req, botToken); err == nil {
		t.Fatal("expected invalid hash error")
	}
}

func TestAuthServiceRefreshRejectsAccessToken(t *testing.T) {
	service, _, user := newTestAuthService(t)

	tokens, err := createTokens(noopLogger{}, config.Config{JWTSecretKey: "test-secret"}, user)
	if err != nil {
		t.Fatalf("failed to create tokens: %v", err)
	}

	_, err = service.Refresh(tokens.AccessToken)
	if !errors.Is(err, appErrors.ErrInvalidRefreshToken) {
		t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestAuthServiceRefreshReturnsNewTokenPairForActiveUser(t *testing.T) {
	service, _, user := newTestAuthService(t)

	tokens, err := createTokens(noopLogger{}, config.Config{JWTSecretKey: "test-secret"}, user)
	if err != nil {
		t.Fatalf("failed to create tokens: %v", err)
	}

	refreshed, err := service.Refresh(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("expected refresh to succeed, got %v", err)
	}
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Fatal("expected both tokens to be returned")
	}
}

func TestAuthServiceMeReturnsCurrentProfile(t *testing.T) {
	service, _, user := newTestAuthService(t)

	view, err := service.Me(user.ID)
	if err != nil {
		t.Fatalf("expected Me to succeed, got %v", err)
	}
	if view.ID != user.ID || view.Username != user.Username || view.Status != user.Status {
		t.Fatalf("unexpected user view: %+v", view)
	}
}

func TestAuthServiceListUsersRejectsInvalidStatus(t *testing.T) {
	service, _, _ := newTestAuthService(t)

	_, err := service.ListUsers("unknown")
	if !errors.Is(err, appErrors.ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestAuthServiceUpdateUserStatusPersistsChange(t *testing.T) {
	service, rep, user := newTestAuthService(t)
	user.Status = "pending"
	if err := rep.Save(user).Error; err != nil {
		t.Fatalf("failed to update initial status: %v", err)
	}

	view, err := service.UpdateUserStatus(user.ID, "active")
	if err != nil {
		t.Fatalf("expected UpdateUserStatus to succeed, got %v", err)
	}
	if view.Status != "active" {
		t.Fatalf("expected response status active, got %s", view.Status)
	}

	reloaded, err := new(models.User).FindByID(rep, int(user.ID))
	if err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if reloaded.Status != "active" {
		t.Fatalf("expected persisted status active, got %s", reloaded.Status)
	}
}

func TestAuthServiceLoginUsesSameErrorForUnknownUsernameAndWrongPassword(t *testing.T) {
	service, _, user := newTestAuthService(t)

	_, missingErr := service.Login("missing-user", "wrong-password")
	if !errors.Is(missingErr, appErrors.ErrInvalidPassword) {
		t.Fatalf("expected missing user to return ErrInvalidPassword, got %v", missingErr)
	}

	_, passwordErr := service.Login(user.Username, "wrong-password")
	if !errors.Is(passwordErr, appErrors.ErrInvalidPassword) {
		t.Fatalf("expected wrong password to return ErrInvalidPassword, got %v", passwordErr)
	}
}

func buildCustomTelegramHash(req dto.TelegramAuthRequest, botToken string) string {
	dataMap := map[string]string{
		"auth_date":         fmt.Sprintf("%d", req.AuthDate),
		"display_name":      req.DisplayName,
		"telegram_id":       fmt.Sprintf("%d", req.TelegramID),
		"telegram_username": req.TelegramUsername,
		"avatar_url":        req.AvatarURL,
	}
	dataCheckString := buildDataCheckStringFromMap(dataMap)
	secret := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secret[:])
	_, _ = mac.Write([]byte(dataCheckString))
	return hex.EncodeToString(mac.Sum(nil))
}

func buildInitDataHash(values url.Values, botToken string) string {
	keys := make([]string, 0, len(values))
	for k := range values {
		if k == "hash" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, k+"="+values.Get(k))
	}
	dataCheckString := strings.Join(lines, "\n")
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	_, _ = secret.Write([]byte(botToken))
	secretKey := secret.Sum(nil)
	mac := hmac.New(sha256.New, secretKey)
	_, _ = mac.Write([]byte(dataCheckString))
	return hex.EncodeToString(mac.Sum(nil))
}

func newTestAuthService(t *testing.T) (AuthServices, repository.Repository, *models.User) {
	t.Helper()

	db := openTestSQLiteDB(t)

	rep := &testRepository{db: db}
	for _, model := range []interface{}{&models.Role{}, &models.User{}} {
		if err := rep.AutoMigrate(model); err != nil {
			t.Fatalf("failed to migrate test db: %v", err)
		}
	}

	for _, role := range []models.Role{{ID: 1, Name: "member"}, {ID: 2, Name: "admin"}} {
		if err := rep.Create(&role).Error; err != nil {
			t.Fatalf("failed to seed role %s: %v", role.Name, err)
		}
	}

	user := &models.User{
		Username:         "tester",
		DisplayName:      "Tester",
		TelegramUsername: "tester_tg",
		AvatarURL:        "https://example.com/avatar.png",
		RoleID:           1,
		Status:           "active",
	}
	var err error
	user.Password, err = utils.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	if err := rep.Create(user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	cont, err := container.NewContainer(rep, noopLogger{}, config.Config{JWTSecretKey: "test-secret"})
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	return NewAuthServices(cont), rep, user
}

func openTestSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()

	id := atomic.AddUint64(&testDBCounter, 1)
	name := fmt.Sprintf("test-%d-%d", time.Now().UnixNano(), id)
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", name)), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	return db
}

type noopLogger struct{}

func (noopLogger) Info(string, ...interface{})  {}
func (noopLogger) Error(string, ...interface{}) {}
func (noopLogger) Debug(string, ...interface{}) {}
func (noopLogger) Warn(string, ...interface{})  {}
func (noopLogger) Fatal(string, ...interface{}) {}
func (noopLogger) Panic(string, ...interface{}) {}

type testRepository struct {
	db *gorm.DB
}

func (r *testRepository) Model(value interface{}) *gorm.DB { return r.db.Model(value) }
func (r *testRepository) Select(query interface{}, args ...interface{}) *gorm.DB {
	return r.db.Select(query, args...)
}
func (r *testRepository) Find(out interface{}, where ...interface{}) *gorm.DB {
	return r.db.Find(out, where...)
}
func (r *testRepository) Exec(query string, values ...interface{}) *gorm.DB {
	return r.db.Exec(query, values...)
}
func (r *testRepository) First(out interface{}, where ...interface{}) *gorm.DB {
	return r.db.First(out, where...)
}
func (r *testRepository) Raw(query string, values ...interface{}) *gorm.DB {
	return r.db.Raw(query, values...)
}
func (r *testRepository) Create(value interface{}) *gorm.DB  { return r.db.Create(value) }
func (r *testRepository) Save(value interface{}) *gorm.DB    { return r.db.Save(value) }
func (r *testRepository) Updates(value interface{}) *gorm.DB { return r.db.Updates(value) }
func (r *testRepository) Delete(value interface{}) *gorm.DB  { return r.db.Delete(value) }
func (r *testRepository) Where(query interface{}, args ...interface{}) *gorm.DB {
	return r.db.Where(query, args...)
}
func (r *testRepository) Preload(column string, conditions ...interface{}) *gorm.DB {
	return r.db.Preload(column, conditions...)
}
func (r *testRepository) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB {
	return r.db.Scopes(funcs...)
}
func (r *testRepository) ScanRows(rows *sql.Rows, result interface{}) error {
	return r.db.ScanRows(rows, result)
}
func (r *testRepository) Clauses(conds ...clause.Expression) *gorm.DB { return r.db.Clauses(conds...) }
func (r *testRepository) AutoMigrate(value interface{}) error         { return r.db.AutoMigrate(value) }
func (r *testRepository) DropTableIfExists(value interface{}) error {
	return r.db.Migrator().DropTable(value)
}
func (r *testRepository) GetSQLDB() (*sql.DB, error) { return r.db.DB() }
func (r *testRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
func (r *testRepository) Transaction(fc func(tx repository.Repository) error) (err error) {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fc(&testRepository{db: tx})
	})
}
