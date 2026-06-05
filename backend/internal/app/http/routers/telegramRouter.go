package routers

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/middlewares"
	"cinema/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterTelegramNotificationRoutes(r *gin.Engine, h handlers.TelegramNotificationHandler, jwtSecret string, reps ...repository.Repository) {
	admin := r.Group("/admin/telegram/notifications")
	admin.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoleNames("admin"))
	{
		admin.GET("", h.List)
		admin.POST("", h.Enqueue)
		admin.POST("/:id/retry", h.Retry)
	}
}
