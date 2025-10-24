package services

import (
	"cinema/internal/app/repository"
	"cinema/internal/app/utils"
	appErrors "cinema/internal/errors"
	"cinema/internal/logger"
	"cinema/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type RegisterServices struct {
	db     *gorm.DB
	logger logger.Logger
}

type AuthServices struct {
	db        *gorm.DB
	logger    logger.Logger
	secretKey string
}

func NewRegisterServices(db *gorm.DB, logger logger.Logger) *RegisterServices {
	return &RegisterServices{
		db:     db,
		logger: logger,
	}
}

func NewAuthServices(db *gorm.DB, logger logger.Logger, secretKey string) *AuthServices {
	return &AuthServices{
		db:        db,
		logger:    logger,
		secretKey: secretKey,
	}
}

func (s *RegisterServices) Register(username, password string) error {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			break
		default:
			s.logger.Error("Error searching user: %s", err)
			return appErrors.ErrUserNameAlreadyExist
		}
	}
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		s.logger.Error("Error hashing password: %s", err)
		return appErrors.ErrInvalidServer
	}
	user = models.User{
		Username: username,
		Password: hashedPassword,
		RoleID:   1,
	}

	err = s.db.Create(&user).Error
	if err != nil {
		s.logger.Error("Error creating user: %s", err)
		return appErrors.ErrInvalidServer
	}
	return nil
}

func (auth *AuthServices) Login(username, password string) (tokenPair repository.TokenPair, err error) {
	var user models.User
	if err = auth.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			auth.logger.Info("User not found", err)
			return tokenPair, appErrors.ErrUserNotFound
		}
		auth.logger.Error("Error searching user", err)
		return tokenPair, appErrors.ErrInvalidServer
	}
	if !utils.VerifyPassword(user.Password, password) {
		return tokenPair, appErrors.ErrInvalidPassword
	}
	tokenPair, err = auth.createTokens(&user)
	if err != nil {
		auth.logger.Error("Error creating JWT", err)
		return tokenPair, appErrors.ErrProblemWithCreateJWT
	}

	return tokenPair, nil
}

// ВСПОМОГАТЕЛЬНОЕ

func (s *AuthServices) createTokens(user *models.User) (tokenPair repository.TokenPair, err error) {
	accessClaims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role_id":  user.RoleID,
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"iat":      time.Now().Unix(),
		"type":     "access",
	}

	access := jwt.NewWithClaims(jwt.SigningMethodHS512, accessClaims)
	tokenPair.AccessToken, err = access.SignedString([]byte(s.secretKey))
	if err != nil {
		s.logger.Error("Error creating JWT access", err)
		return tokenPair, err
	}

	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshClaims)
	tokenPair.RefreshToken, err = refresh.SignedString([]byte(s.secretKey))
	if err != nil {
		s.logger.Error("Error creating JWT refresh", err)
		return tokenPair, err
	}

	return tokenPair, nil
}
