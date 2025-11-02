package routers

import (
	"cinema/internal/app/http/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine, authHandler handlers.AuthHandler) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}
}

func RegisterUserRoutes(r *gin.Engine, userHandler handlers.UserHandler) {
	auth := r.Group("/user")
	{
		auth.POST("/update", userHandler.UpdateUser)
	}
}
