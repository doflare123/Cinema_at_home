package routers

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/middlewares"
	"cinema/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterFranchiseRoutes(r *gin.Engine, franchiseHandler handlers.FranchiseHandler, jwtSecret string, reps ...repository.Repository) {
	r.GET("/franchises", franchiseHandler.List)
	r.GET("/franchises/:id", franchiseHandler.GetByID)

	admin := r.Group("/admin/franchises")
	admin.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoles(2))
	{
		admin.POST("", franchiseHandler.Create)
		admin.POST("/:id/movies", franchiseHandler.AddMovie)
	}
}
