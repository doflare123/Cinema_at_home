package routers

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/middlewares"
	"cinema/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine, authHandler handlers.AuthHandler, reps ...repository.Repository) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/telegram", authHandler.Telegram)
		auth.POST("/refresh", authHandler.Refresh)
	}

	authProtected := auth.Group("")
	authProtected.Use(middlewares.JWTAuthMiddleware(authHandler.JWTSecret(), reps...), middlewares.RequireActiveStatus())
	{
		authProtected.GET("/me", authHandler.Me)
	}

	admin := r.Group("/admin")
	admin.Use(middlewares.JWTAuthMiddleware(authHandler.JWTSecret(), reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoleNames("admin"))
	{
		admin.GET("/users", authHandler.ListUsers)
		admin.PATCH("/users/:id/status", authHandler.UpdateUserStatus)
	}
}

func RegisterUserRoutes(r *gin.Engine, userHandler handlers.UserHandler, jwtSecret string, reps ...repository.Repository) {
	group := r.Group("/user")
	group.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus())
	{
		group.POST("/update", userHandler.UpdateUser)
	}
}
