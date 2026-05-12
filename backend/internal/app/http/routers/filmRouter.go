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

	films := r.Group("/film")
	films.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoles(1, 2))
	{
		films.POST("/", filmH.CreateFilm)
	}
}
