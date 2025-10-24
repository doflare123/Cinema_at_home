package handlers

import (
	"cinema/internal/app/repository"
	"cinema/internal/app/services"
	"cinema/internal/app/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RegisterHandler struct {
	AuthServices *services.RegisterServices
}

type AuthHandler struct {
	AuthServices *services.AuthServices
}

func NewAuthHandler(authServices *services.AuthServices) *AuthHandler {
	return &AuthHandler{
		AuthServices: authServices,
	}
}

func NewRegisterHandler(authServices *services.RegisterServices) *RegisterHandler {
	return &RegisterHandler{
		AuthServices: authServices,
	}
}

func (h *RegisterHandler) Register(c *gin.Context) {
	var request repository.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		if utils.HandleValidationError(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.AuthServices.Register(request.Username, request.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request repository.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		if utils.HandleValidationError(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := h.AuthServices.Login(request.Username, request.Password)
	if err != nil {
		fmt.Print(err)
		if utils.HandleValidationError(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	})
}
