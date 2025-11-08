package routers

import (
	"cinema/internal/app/http/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterFilmRoutes(r *gin.Engine, filmH handlers.FilmHandler) {
	films := r.Group("/film")
	{
		films.POST("/", filmH.CreateFilm)
	}
}
