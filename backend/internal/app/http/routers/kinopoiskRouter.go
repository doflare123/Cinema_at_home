package routers

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/middlewares"
	"cinema/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterKinopoiskRoutes(r *gin.Engine, h handlers.KinopoiskHandler, jwtSecret string, reps ...repository.Repository) {
	kinopoisk := r.Group("/kinopoisk")
	kinopoisk.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoleNames("member", "admin"))
	{
		kinopoisk.GET("/search", h.Search)
	}
}
