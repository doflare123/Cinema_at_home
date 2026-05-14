package routers

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/middlewares"
	"cinema/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterMovieProposalRoutes(r *gin.Engine, h handlers.MovieProposalHandler, jwtSecret string, reps ...repository.Repository) {
	proposals := r.Group("/proposals")
	proposals.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoleNames("member", "admin"))
	{
		proposals.POST("", h.Create)
		proposals.GET("/me", h.My)
	}

	admin := r.Group("/admin/proposals")
	admin.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoleNames("admin"))
	{
		admin.GET("", h.List)
		admin.PATCH("/:id/status", h.Moderate)
	}
}
