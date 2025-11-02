package handlers

import (
	"cinema/internal/app/services"
	"cinema/internal/container"
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
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пользователь успешно обновлён"})
}
