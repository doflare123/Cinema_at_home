package routers

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/middlewares"
	"cinema/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterFilmRoutes(r *gin.Engine, filmH handlers.FilmHandler, jwtSecret string, reps ...repository.Repository) {
	r.GET("/movies", filmH.List)
	r.GET("/movies/:id", filmH.GetByID)

	adminFilms := r.Group("/admin/movies")
	adminFilms.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoleNames("admin"))
	{
		adminFilms.POST("", filmH.CreateFilm)
	}

	// Legacy alias kept for backward compatibility.
	legacyFilms := r.Group("/film")
	legacyFilms.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoleNames("admin"))
	{
		legacyFilms.POST("/", filmH.CreateFilm)
	}
}
