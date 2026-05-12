package handlers

import (
	"bytes"
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubAuthService struct {
	loginTelegram func(req dto.TelegramAuthRequest) (dto.TokenPair, *models.User, error)
}

func (s stubAuthService) Register(username, password, displayName string) error {
	return nil
}

func (s stubAuthService) Login(username, password string) (dto.TokenPair, error) {
	return dto.TokenPair{}, nil
}

func (s stubAuthService) LoginTelegram(req dto.TelegramAuthRequest) (dto.TokenPair, *models.User, error) {
	if s.loginTelegram != nil {
		return s.loginTelegram(req)
	}
	return dto.TokenPair{}, nil, nil
}

func (s stubAuthService) Refresh(refreshToken string) (dto.TokenPair, error) {
	return dto.TokenPair{}, nil
}

func (s stubAuthService) Me(userID uint) (dto.UserView, error) {
	return dto.UserView{}, nil
}

func (s stubAuthService) ListUsers(status string) ([]dto.UserView, error) {
	return nil, nil
}

func (s stubAuthService) UpdateUserStatus(userID uint, status string) (dto.UserView, error) {
	return dto.UserView{}, nil
}

func TestAuthHandlerTelegramReturnsActualInactiveStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &authHandler{
		service: stubAuthService{
			loginTelegram: func(req dto.TelegramAuthRequest) (dto.TokenPair, *models.User, error) {
				return dto.TokenPair{}, &models.User{
					DisplayName: "Blocked User",
					Status:      "blocked",
				}, appErrors.ErrUserNotActive
			},
		},
	}

	body, err := json.Marshal(dto.TelegramAuthRequest{
		TelegramID:  42,
		DisplayName: "Blocked User",
	})
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/auth/telegram", bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Telegram(ctx)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", recorder.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "blocked" {
		t.Fatalf("expected status blocked, got %v", response["status"])
	}
}
