package routers

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine, authHandler handlers.AuthHandler) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/telegram", authHandler.Telegram)
	}
}

func RegisterUserRoutes(r *gin.Engine, userHandler handlers.UserHandler, jwtSecret string) {
	group := r.Group("/user")
	group.Use(middlewares.JWTAuthMiddleware(jwtSecret), middlewares.RequireActiveStatus())
	{
		group.POST("/update", userHandler.UpdateUser)
	}
}
