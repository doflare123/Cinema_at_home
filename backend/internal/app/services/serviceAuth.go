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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthServices interface {
	Register(username, password, displayName string) error
	Login(username, password string) (dto.TokenPair, error)
	LoginTelegram(req dto.TelegramAuthRequest) (dto.TokenPair, *models.User, error)
	Refresh(refreshToken string) (dto.TokenPair, error)
	Me(userID uint) (dto.UserView, error)
	ListUsers(status string) ([]dto.UserView, error)
	UpdateUserStatus(userID uint, status string) (dto.UserView, error)
}

type authServices struct {
	container container.Container
}

type telegramProfile struct {
	TelegramID       int64
	TelegramUsername string
	DisplayName      string
	AvatarURL        string
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
		return tokenPair, appErrors.ErrInvalidPassword
	}
	if !utils.VerifyPassword(user.Password, password) {
		return tokenPair, appErrors.ErrInvalidPassword
	}
	if user.Status != "active" {
		return tokenPair, appErrors.ErrUserNotActive
	}

	tokenPair, err = createTokens(s.container.GetLogger(), *s.container.GetConfig(), user)
	if err != nil {
		return tokenPair, appErrors.ErrProblemWithCreateJWT
	}

	return tokenPair, nil
}

func (s *authServices) LoginTelegram(req dto.TelegramAuthRequest) (tokenPair dto.TokenPair, user *models.User, err error) {
	profile, err := verifyAndExtractTelegram(req, s.container.GetConfig().TelegramBotToken)
	if err != nil {
		return tokenPair, nil, err
	}

	user, err = new(models.User).FindByTelegramID(s.container.GetRepository(), profile.TelegramID)
	if err != nil {
		if !errors.Is(err, appErrors.ErrUserNotFound) {
			return tokenPair, nil, appErrors.ErrInvalidServer
		}

		username := profile.TelegramUsername
		if username == "" {
			username = fmt.Sprintf("tg_%d", profile.TelegramID)
		}
		memberRoleID, roleErr := getRoleIDByName(s.container, "member")
		if roleErr != nil {
			return tokenPair, nil, appErrors.ErrInvalidServer
		}

		user = &models.User{
			TelegramID:       &profile.TelegramID,
			TelegramUsername: profile.TelegramUsername,
			DisplayName:      profile.DisplayName,
			AvatarURL:        profile.AvatarURL,
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

	if user.Status != "active" {
		return tokenPair, user, appErrors.ErrUserNotActive
	}

	updates := map[string]interface{}{
		"telegram_username": profile.TelegramUsername,
		"display_name":      profile.DisplayName,
		"avatar_url":        profile.AvatarURL,
	}
	if err := user.Update(s.container.GetRepository(), updates); err != nil {
		return tokenPair, nil, appErrors.ErrUpdDataUser
	}

	tokenPair, err = createTokens(s.container.GetLogger(), *s.container.GetConfig(), user)
	if err != nil {
		return tokenPair, nil, appErrors.ErrProblemWithCreateJWT
	}
	return tokenPair, user, nil
}

func (s *authServices) Refresh(refreshToken string) (dto.TokenPair, error) {
	var out dto.TokenPair
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.container.GetConfig().JWTSecretKey), nil
	})
	if err != nil || !token.Valid {
		return out, appErrors.ErrInvalidRefreshToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return out, appErrors.ErrInvalidRefreshToken
	}
	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return out, appErrors.ErrInvalidRefreshToken
	}
	userID, ok := getUintClaim(claims, "user_id")
	if !ok {
		return out, appErrors.ErrInvalidRefreshToken
	}
	user, err := new(models.User).FindByID(s.container.GetRepository(), int(userID))
	if err != nil {
		return out, appErrors.ErrUserNotFound
	}
	if user.Status != "active" {
		return out, appErrors.ErrUserNotActive
	}
	return createTokens(s.container.GetLogger(), *s.container.GetConfig(), user)
}

func (s *authServices) Me(userID uint) (dto.UserView, error) {
	user, err := new(models.User).FindByID(s.container.GetRepository(), int(userID))
	if err != nil {
		return dto.UserView{}, appErrors.ErrUserNotFound
	}
	return buildUserView(user), nil
}

func (s *authServices) ListUsers(status string) ([]dto.UserView, error) {
	normalizedStatus, err := normalizeUserStatus(status, true)
	if err != nil {
		return nil, err
	}

	users, err := new(models.User).ListByStatus(s.container.GetRepository(), normalizedStatus)
	if err != nil {
		return nil, appErrors.ErrInvalidServer
	}

	out := make([]dto.UserView, 0, len(users))
	for i := range users {
		out = append(out, buildUserView(&users[i]))
	}
	return out, nil
}

func (s *authServices) UpdateUserStatus(userID uint, status string) (dto.UserView, error) {
	normalizedStatus, err := normalizeUserStatus(status, false)
	if err != nil {
		return dto.UserView{}, err
	}

	user, err := new(models.User).FindByID(s.container.GetRepository(), int(userID))
	if err != nil {
		return dto.UserView{}, appErrors.ErrUserNotFound
	}

	if user.Status == normalizedStatus {
		return buildUserView(user), nil
	}

	if err := user.Update(s.container.GetRepository(), map[string]interface{}{"status": normalizedStatus}); err != nil {
		return dto.UserView{}, appErrors.ErrInvalidServer
	}
	user.Status = normalizedStatus
	return buildUserView(user), nil
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

func verifyAndExtractTelegram(req dto.TelegramAuthRequest, botToken string) (telegramProfile, error) {
	if botToken == "" {
		return telegramProfile{}, appErrors.ErrInvalidServer
	}

	if strings.TrimSpace(req.InitData) != "" {
		return verifyAndExtractTelegramInitData(req.InitData, botToken)
	}

	if err := verifyTelegramAuthPayload(req, botToken); err != nil {
		return telegramProfile{}, err
	}
	if req.TelegramID == 0 || req.DisplayName == "" {
		return telegramProfile{}, appErrors.ErrInvalidTelegramAuth
	}

	return telegramProfile{
		TelegramID:       req.TelegramID,
		TelegramUsername: req.TelegramUsername,
		DisplayName:      req.DisplayName,
		AvatarURL:        req.AvatarURL,
	}, nil
}

func verifyAndExtractTelegramInitData(initData, botToken string) (telegramProfile, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return telegramProfile{}, appErrors.ErrInvalidTelegramAuth
	}

	hash := values.Get("hash")
	authDate := values.Get("auth_date")
	if hash == "" || authDate == "" {
		return telegramProfile{}, appErrors.ErrInvalidTelegramAuth
	}
	if !verifyTelegramDataCheckString(values, hash, botToken) {
		return telegramProfile{}, appErrors.ErrInvalidTelegramAuth
	}

	authUnix, err := strconv.ParseInt(authDate, 10, 64)
	if err != nil {
		return telegramProfile{}, appErrors.ErrInvalidTelegramAuth
	}
	if isAuthDateInvalid(authUnix) {
		return telegramProfile{}, appErrors.ErrInvalidTelegramAuth
	}

	if userRaw := values.Get("user"); userRaw != "" {
		var userData struct {
			ID        int64  `json:"id"`
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			PhotoURL  string `json:"photo_url"`
		}
		if err := json.Unmarshal([]byte(userRaw), &userData); err != nil {
			return telegramProfile{}, appErrors.ErrInvalidTelegramAuth
		}
		if userData.ID == 0 {
			return telegramProfile{}, appErrors.ErrInvalidTelegramAuth
		}
		displayName := strings.TrimSpace(userData.FirstName + " " + userData.LastName)
		if displayName == "" {
			displayName = userData.Username
		}
		if displayName == "" {
			displayName = fmt.Sprintf("tg_%d", userData.ID)
		}
		return telegramProfile{
			TelegramID:       userData.ID,
			TelegramUsername: userData.Username,
			DisplayName:      displayName,
			AvatarURL:        userData.PhotoURL,
		}, nil
	}

	tgID, err := strconv.ParseInt(values.Get("id"), 10, 64)
	if err != nil || tgID == 0 {
		return telegramProfile{}, appErrors.ErrInvalidTelegramAuth
	}
	displayName := strings.TrimSpace(values.Get("first_name") + " " + values.Get("last_name"))
	if displayName == "" {
		displayName = values.Get("username")
	}
	if displayName == "" {
		displayName = fmt.Sprintf("tg_%d", tgID)
	}
	return telegramProfile{
		TelegramID:       tgID,
		TelegramUsername: values.Get("username"),
		DisplayName:      displayName,
		AvatarURL:        values.Get("photo_url"),
	}, nil
}

func verifyTelegramAuthPayload(req dto.TelegramAuthRequest, botToken string) error {
	if req.AuthDate <= 0 {
		return appErrors.ErrInvalidTelegramAuth
	}
	if isAuthDateInvalid(req.AuthDate) {
		return appErrors.ErrInvalidTelegramAuth
	}

	dataMap := map[string]string{
		"auth_date":         fmt.Sprintf("%d", req.AuthDate),
		"display_name":      req.DisplayName,
		"telegram_id":       fmt.Sprintf("%d", req.TelegramID),
		"telegram_username": req.TelegramUsername,
		"avatar_url":        req.AvatarURL,
	}
	dataCheckString := buildDataCheckStringFromMap(dataMap)
	if !verifyTelegramHash(dataCheckString, req.Hash, botToken) {
		return appErrors.ErrInvalidTelegramAuth
	}
	return nil
}

func verifyTelegramDataCheckString(values url.Values, providedHash, botToken string) bool {
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
		v := values.Get(k)
		if v == "" {
			continue
		}
		lines = append(lines, k+"="+v)
	}
	dataCheckString := strings.Join(lines, "\n")
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	_, _ = secret.Write([]byte(botToken))
	secretKey := secret.Sum(nil)

	mac := hmac.New(sha256.New, secretKey)
	_, _ = mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.ToLower(expectedHash)), []byte(strings.ToLower(providedHash)))
}

func verifyTelegramHash(dataCheckString, providedHash, botToken string) bool {
	secret := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secret[:])
	_, _ = mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.ToLower(expectedHash)), []byte(strings.ToLower(providedHash)))
}

func buildDataCheckStringFromMap(dataMap map[string]string) string {
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
	return strings.Join(lines, "\n")
}

func isAuthDateInvalid(authDate int64) bool {
	now := time.Now().Unix()
	if authDate > now+300 {
		return true
	}
	if now-authDate > 86400 {
		return true
	}
	return false
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

func getUintClaim(claims jwt.MapClaims, key string) (uint, bool) {
	val, ok := claims[key]
	if !ok {
		return 0, false
	}
	floatVal, ok := val.(float64)
	if !ok || floatVal < 0 {
		return 0, false
	}
	return uint(floatVal), true
}

func buildUserView(user *models.User) dto.UserView {
	return dto.UserView{
		ID:               user.ID,
		Username:         user.Username,
		DisplayName:      user.DisplayName,
		TelegramUsername: user.TelegramUsername,
		AvatarURL:        user.AvatarURL,
		RoleID:           user.RoleID,
		Status:           user.Status,
	}
}

func normalizeUserStatus(status string, allowEmpty bool) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(status))
	if normalized == "" {
		if allowEmpty {
			return "", nil
		}
		return "", appErrors.ErrInvalidStatus
	}

	switch normalized {
	case "pending", "active", "rejected", "blocked":
		return normalized, nil
	default:
		return "", appErrors.ErrInvalidStatus
	}
}
