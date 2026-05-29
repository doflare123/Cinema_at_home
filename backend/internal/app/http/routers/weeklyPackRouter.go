package routers

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/middlewares"
	"cinema/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterWeeklyPackRoutes(r *gin.Engine, h handlers.WeeklyPackHandler, jwtSecret string, reps ...repository.Repository) {
	r.GET("/weekly-packs", h.List)
	r.GET("/weekly-packs/current", h.Current)
	r.GET("/weekly-packs/:id", h.GetByID)

	votes := r.Group("/weekly-packs")
	votes.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoleNames("member", "admin"))
	{
		votes.POST("/:id/votes", h.UpsertVote)
		votes.GET("/:id/votes/me", h.MeVotes)
		votes.GET("/:id/votes/me/limits", h.MeVoteLimits)
	}

	admin := r.Group("/admin/weekly-packs")
	admin.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoleNames("admin"))
	{
		admin.POST("", h.Create)
		admin.POST("/:id/movies", h.AddMovie)
		admin.PATCH("/:id/status", h.UpdateStatus)
	}
}
