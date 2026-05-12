package routers

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/middlewares"
	"cinema/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterReviewRoutes(r *gin.Engine, h handlers.ReviewHandler, jwtSecret string, reps ...repository.Repository) {
	r.GET("/reviews/films/:filmId", h.ListByFilm)

	reviews := r.Group("/reviews")
	reviews.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoleNames("member", "admin"))
	{
		reviews.GET("/films/:filmId/me", h.Me)
		reviews.POST("/films/:filmId", h.Upsert)
	}
}
