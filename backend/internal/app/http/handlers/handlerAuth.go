package handlers

import (
	"cinema/internal/app/services"
	"cinema/internal/app/utils"
	"cinema/internal/container"
	appErrors "cinema/internal/errors"
	"cinema/internal/models/dto"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Telegram(c *gin.Context)
	Refresh(c *gin.Context)
	Me(c *gin.Context)
	ListUsers(c *gin.Context)
	UpdateUserStatus(c *gin.Context)
	JWTSecret() string
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

func (h *authHandler) JWTSecret() string {
	return h.container.GetConfig().JWTSecretKey
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
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
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
			status := "pending"
			message := "Account is not active"
			displayName := ""
			if user != nil {
				if user.Status != "" {
					status = user.Status
				}
				displayName = user.DisplayName
			}
			switch status {
			case "pending":
				message = "Account awaits admin approval"
			case "rejected":
				message = "Account was rejected by admin"
			case "blocked":
				message = "Account is blocked"
			}
			c.JSON(http.StatusAccepted, gin.H{
				"status":       status,
				"message":      message,
				"display_name": displayName,
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

func (h *authHandler) Refresh(c *gin.Context) {
	var request dto.RefreshRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		if utils.HandleValidationError(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.service.Refresh(request.RefreshToken)
	if err != nil {
		switch err {
		case appErrors.ErrInvalidRefreshToken:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case appErrors.ErrUserNotActive:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case appErrors.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func (h *authHandler) Me(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	user, err := h.service.Me(userID)
	if err != nil {
		if err == appErrors.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *authHandler) ListUsers(c *gin.Context) {
	users, err := h.service.ListUsers(c.Query("status"))
	if err != nil {
		if err == appErrors.ErrInvalidStatus {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *authHandler) UpdateUserStatus(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var request dto.AdminUpdateUserStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		if utils.HandleValidationError(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, serviceErr := h.service.UpdateUserStatus(uint(userID), request.Status)
	if serviceErr != nil {
		switch serviceErr {
		case appErrors.ErrInvalidStatus:
			c.JSON(http.StatusBadRequest, gin.H{"error": serviceErr.Error()})
		case appErrors.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": serviceErr.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
