package services

import (
	"cinema/config"
	"cinema/internal/app/utils"
	"cinema/internal/container"
	appErrors "cinema/internal/errors"
	"cinema/internal/logger"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthServices interface {
	Register(username, password string) error
	Login(username, password string) (dto.TokenPair, error)
}

type authServices struct {
	container container.Container
}

func NewAuthServices(container container.Container) AuthServices {
	return &authServices{
		container: container,
	}
}

func (s *authServices) Register(username, password string) error {
	var user models.User
	err := user.NameAlreadyExist(s.container.GetRepository(), username)
	if err == nil {
		s.container.GetLogger().Info("User already exist", "username", username)
		return appErrors.ErrUserNameAlreadyExist
	}
	if err != gorm.ErrRecordNotFound {
		return appErrors.ErrInvalidServer
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		s.container.GetLogger().Error("Error hashing password", "error", err.Error())
		return appErrors.ErrInvalidServer
	}
	user = models.User{
		Username: username,
		Password: hashedPassword,
		RoleID:   1,
	}

	err = user.Create(s.container.GetRepository())
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			break
		case appErrors.ErrUserNameAlreadyExist:
			s.container.GetLogger().Error("Error searching user", "error", err.Error())
			return err
		}
		s.container.GetLogger().Error("Error creating user", "error", err.Error())
		return appErrors.ErrInvalidServer
	}
	return nil
}

func (s *authServices) Login(username, password string) (tokenPair dto.TokenPair, err error) {
	user, err := new(models.User).FindByName(s.container.GetRepository(), username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.container.GetLogger().Info("User not found", err.Error())
			return tokenPair, appErrors.ErrUserNotFound
		}
		s.container.GetLogger().Error("Error searching user", err.Error())
		return tokenPair, appErrors.ErrInvalidServer
	}
	if !utils.VerifyPassword(user.Password, password) {
		return tokenPair, appErrors.ErrInvalidPassword
	}
	tokenPair, err = createTokens(s.container.GetLogger(), *s.container.GetConfig(), user)
	if err != nil {
		s.container.GetLogger().Error("Error creating JWT", err.Error())
		return tokenPair, appErrors.ErrProblemWithCreateJWT
	}

	return tokenPair, nil
}

// ВСПОМОГАТЕЛЬНОЕ

func createTokens(logger logger.Logger, config config.Config, user *models.User) (tokenPair dto.TokenPair, err error) {
	secret := config.JWTSecretKey
	accessClaims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role_id":  user.RoleID,
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
