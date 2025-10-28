package services

import (
	"cinema/internal/app/utils"
	"cinema/internal/container"
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type UserServices interface {
	Register(username, password string) error
	Login(username, password string) (dto.TokenPair, error)
}

type userServices struct {
	container container.Container
}

func NewUserServices(container container.Container) UserServices {
	return &userServices{
		container: container,
	}
}

func (s *userServices) Register(username, password string) error {
	var user models.User
	if err := s.container.GetRepository().Where("username = ?", username).First(&user).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			break
		default:
			s.container.GetLogger().Error("Error searching user: %s", err)
			return appErrors.ErrUserNameAlreadyExist
		}
	}
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		s.container.GetLogger().Error("Error hashing password: %s", err)
		return appErrors.ErrInvalidServer
	}
	user = models.User{
		Username: username,
		Password: hashedPassword,
		RoleID:   1,
	}

	err = s.container.GetRepository().Create(&user).Error
	if err != nil {
		s.container.GetLogger().Error("Error creating user: %s", err)
		return appErrors.ErrInvalidServer
	}
	return nil
}

func (s *userServices) Login(username, password string) (tokenPair dto.TokenPair, err error) {
	var user models.User
	if err = s.container.GetRepository().Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.container.GetLogger().Info("User not found", err)
			return tokenPair, appErrors.ErrUserNotFound
		}
		s.container.GetLogger().Error("Error searching user", err)
		return tokenPair, appErrors.ErrInvalidServer
	}
	if !utils.VerifyPassword(user.Password, password) {
		return tokenPair, appErrors.ErrInvalidPassword
	}
	tokenPair, err = s.createTokens(&user)
	if err != nil {
		s.container.GetLogger().Error("Error creating JWT", err)
		return tokenPair, appErrors.ErrProblemWithCreateJWT
	}

	return tokenPair, nil
}

// ВСПОМОГАТЕЛЬНОЕ

func (s *userServices) createTokens(user *models.User) (tokenPair dto.TokenPair, err error) {
	accessClaims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role_id":  user.RoleID,
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"iat":      time.Now().Unix(),
		"type":     "access",
	}

	access := jwt.NewWithClaims(jwt.SigningMethodHS512, accessClaims)
	tokenPair.AccessToken, err = access.SignedString([]byte(s.container.GetConfig().JWTSecretKey))
	if err != nil {
		s.container.GetLogger().Error("Error creating JWT access", err)
		return tokenPair, err
	}

	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshClaims)
	tokenPair.RefreshToken, err = refresh.SignedString([]byte(s.container.GetConfig().JWTSecretKey))
	if err != nil {
		s.container.GetLogger().Error("Error creating JWT refresh", err)
		return tokenPair, err
	}

	return tokenPair, nil
}
