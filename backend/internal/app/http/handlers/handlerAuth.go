package handlers

import (
	"cinema/internal/app/services"
	"cinema/internal/app/utils"
	"cinema/internal/container"
	appErrors "cinema/internal/errors"
	"cinema/internal/models/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Telegram(c *gin.Context)
}

type authHandler struct {
	container container.Container
	service   services.AuthServices
}

func NewAuthHandler(container container.Container) AuthHandler {
	return &authHandler{
		container: container,
		service:   services.NewAuthServices(container),
	}
}

func (h *authHandler) Register(c *gin.Context) {
	var request dto.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		if utils.HandleValidationError(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Register(request.Username, request.Password, request.Display); err != nil {
		if err == appErrors.ErrUserNameAlreadyExist {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered. Await admin approval."})
}

func (h *authHandler) Login(c *gin.Context) {
	var request dto.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		if utils.HandleValidationError(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := h.service.Login(request.Username, request.Password)
	if err != nil {
		status := http.StatusBadRequest
		if err == appErrors.ErrUserNotActive {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	})
}

func (h *authHandler) Telegram(c *gin.Context) {
	var request dto.TelegramAuthRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		if utils.HandleValidationError(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, user, err := h.service.LoginTelegram(request)
	if err != nil {
		if err == appErrors.ErrInvalidTelegramAuth {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if err == appErrors.ErrUserNameAlreadyExist {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err == appErrors.ErrUserNotActive {
			c.JSON(http.StatusAccepted, gin.H{
				"status":       "pending",
				"message":      "Account awaits admin approval",
				"display_name": user.DisplayName,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}
