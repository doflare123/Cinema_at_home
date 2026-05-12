package routers

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/middlewares"
	"cinema/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterExpectationRoutes(r *gin.Engine, expectationHandler handlers.ExpectationHandler, jwtSecret string, reps ...repository.Repository) {
	r.GET("/expectations/:targetType/:id/summary", expectationHandler.Summary)

	expectations := r.Group("/expectations")
	expectations.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoles(1, 2))
	{
		expectations.POST("", expectationHandler.Upsert)
		expectations.GET("/me", expectationHandler.Me)
	}
}
