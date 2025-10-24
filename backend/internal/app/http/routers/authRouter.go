package routers

import (
	"cinema/internal/app/http/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine, RegisterHandler *handlers.RegisterHandler, AuthHandler *handlers.AuthHandler) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", RegisterHandler.Register)
		auth.POST("/login", AuthHandler.Login)
	}
}
