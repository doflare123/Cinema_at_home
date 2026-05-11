package routers

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterFilmRoutes(r *gin.Engine, filmH handlers.FilmHandler, jwtSecret string) {
	films := r.Group("/film")
	films.Use(middlewares.JWTAuthMiddleware(jwtSecret), middlewares.RequireActiveStatus(), middlewares.RequireRoles(1, 2))
	{
		films.POST("/", filmH.CreateFilm)
	}
}
