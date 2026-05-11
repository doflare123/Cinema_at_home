package services

import (
	"cinema/config"
	"cinema/internal/app/utils"
	"cinema/internal/container"
	appErrors "cinema/internal/errors"
	"cinema/internal/logger"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthServices interface {
	Register(username, password, displayName string) error
	Login(username, password string) (dto.TokenPair, error)
	LoginTelegram(req dto.TelegramAuthRequest) (dto.TokenPair, *models.User, error)
}

type authServices struct {
	container container.Container
}

func NewAuthServices(container container.Container) AuthServices {
	return &authServices{container: container}
}

func (s *authServices) Register(username, password, displayName string) error {
	var user models.User
	err := user.NameAlreadyExist(s.container.GetRepository(), username)
	if err == nil {
		return appErrors.ErrUserNameAlreadyExist
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.ErrInvalidServer
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return appErrors.ErrInvalidServer
	}
	if displayName == "" {
		displayName = username
	}

	memberRoleID, err := getRoleIDByName(s.container, "member")
	if err != nil {
		return appErrors.ErrInvalidServer
	}

	user = models.User{
		Username:    username,
		DisplayName: displayName,
		Password:    hashedPassword,
		RoleID:      memberRoleID,
		Status:      "pending",
	}

	if err := user.Create(s.container.GetRepository()); err != nil {
		if isUniqueViolation(err) {
			return appErrors.ErrUserNameAlreadyExist
		}
		return appErrors.ErrInvalidServer
	}
	return nil
}

func (s *authServices) Login(username, password string) (tokenPair dto.TokenPair, err error) {
	user, err := new(models.User).FindByName(s.container.GetRepository(), username)
	if err != nil {
		return tokenPair, appErrors.ErrUserNotFound
	}
	if user.Status != "active" {
		return tokenPair, appErrors.ErrUserNotActive
	}
	if !utils.VerifyPassword(user.Password, password) {
		return tokenPair, appErrors.ErrInvalidPassword
	}

	tokenPair, err = createTokens(s.container.GetLogger(), *s.container.GetConfig(), user)
	if err != nil {
		return tokenPair, appErrors.ErrProblemWithCreateJWT
	}

	return tokenPair, nil
}

func (s *authServices) LoginTelegram(req dto.TelegramAuthRequest) (tokenPair dto.TokenPair, user *models.User, err error) {
	if err := verifyTelegramAuthPayload(req, s.container.GetConfig().TelegramBotToken); err != nil {
		return tokenPair, nil, err
	}

	user, err = new(models.User).FindByTelegramID(s.container.GetRepository(), req.TelegramID)
	if err != nil {
		if !errors.Is(err, appErrors.ErrUserNotFound) {
			return tokenPair, nil, appErrors.ErrInvalidServer
		}

		username := req.TelegramUsername
		if username == "" {
			username = fmt.Sprintf("tg_%d", req.TelegramID)
		}
		memberRoleID, roleErr := getRoleIDByName(s.container, "member")
		if roleErr != nil {
			return tokenPair, nil, appErrors.ErrInvalidServer
		}

		user = &models.User{
			TelegramID:       &req.TelegramID,
			TelegramUsername: req.TelegramUsername,
			DisplayName:      req.DisplayName,
			AvatarURL:        req.AvatarURL,
			Username:         username,
			Password:         "telegram_auth",
			RoleID:           memberRoleID,
			Status:           "pending",
		}
		if err := user.Create(s.container.GetRepository()); err != nil {
			if isUniqueViolation(err) {
				return tokenPair, nil, appErrors.ErrUserNameAlreadyExist
			}
			return tokenPair, nil, appErrors.ErrInvalidServer
		}
		return tokenPair, user, appErrors.ErrUserNotActive
	}

	updates := map[string]interface{}{
		"telegram_username": req.TelegramUsername,
		"display_name":      req.DisplayName,
		"avatar_url":        req.AvatarURL,
	}
	if err := user.Update(s.container.GetRepository(), updates); err != nil {
		return tokenPair, nil, appErrors.ErrUpdDataUser
	}

	if user.Status != "active" {
		return tokenPair, user, appErrors.ErrUserNotActive
	}

	tokenPair, err = createTokens(s.container.GetLogger(), *s.container.GetConfig(), user)
	if err != nil {
		return tokenPair, nil, appErrors.ErrProblemWithCreateJWT
	}
	return tokenPair, user, nil
}

func createTokens(logger logger.Logger, config config.Config, user *models.User) (tokenPair dto.TokenPair, err error) {
	secret := config.JWTSecretKey
	accessClaims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role_id":  user.RoleID,
		"status":   user.Status,
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"iat":      time.Now().Unix(),
		"type":     "access",
	}

	access := jwt.NewWithClaims(jwt.SigningMethodHS512, accessClaims)
	tokenPair.AccessToken, err = access.SignedString([]byte(secret))
	if err != nil {
		logger.Error("Error creating JWT access", err)
		return tokenPair, err
	}

	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"status":  user.Status,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshClaims)
	tokenPair.RefreshToken, err = refresh.SignedString([]byte(secret))
	if err != nil {
		logger.Error("Error creating JWT refresh", err)
		return tokenPair, err
	}

	return tokenPair, nil
}

func verifyTelegramAuthPayload(req dto.TelegramAuthRequest, botToken string) error {
	if botToken == "" {
		return appErrors.ErrInvalidServer
	}
	if req.AuthDate <= 0 {
		return appErrors.ErrInvalidTelegramAuth
	}
	if time.Now().Unix()-req.AuthDate > 86400 {
		return appErrors.ErrInvalidTelegramAuth
	}

	dataMap := map[string]string{
		"auth_date":         fmt.Sprintf("%d", req.AuthDate),
		"display_name":      req.DisplayName,
		"telegram_id":       fmt.Sprintf("%d", req.TelegramID),
		"telegram_username": req.TelegramUsername,
		"avatar_url":        req.AvatarURL,
	}
	keys := make([]string, 0, len(dataMap))
	for k, v := range dataMap {
		if v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, k+"="+dataMap[k])
	}
	dataCheckString := strings.Join(lines, "\n")

	secret := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secret[:])
	_, _ = mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(strings.ToLower(expectedHash)), []byte(strings.ToLower(req.Hash))) {
		return appErrors.ErrInvalidTelegramAuth
	}
	return nil
}

func getRoleIDByName(cont container.Container, roleName string) (uint, error) {
	role, err := new(models.Role).FindByName(cont.GetRepository(), roleName)
	if err != nil {
		return 0, err
	}
	return role.ID, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
