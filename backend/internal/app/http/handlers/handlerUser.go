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

type UserHandler interface {
	UpdateUser(c *gin.Context)
}

type userHandler struct {
	container container.Container
	service   services.UserServices
}

func NewUserHandler(container container.Container) UserHandler {
	return &userHandler{
		container: container,
		service:   services.NewUserService(container),
	}
}

func (h *userHandler) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if utils.HandleValidationError(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

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

	tokens, err := h.service.UpdateInf(userID, req.Username, req.Password, req.DisplayName, req.AvatarURL)
	if err != nil {
		if err == appErrors.ErrNotEnougthData {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err == appErrors.ErrUserNameAlreadyExist {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "user updated",
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}
