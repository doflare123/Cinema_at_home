package routers

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/middlewares"
	"cinema/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterImportRoutes(r *gin.Engine, h handlers.ImportHandler, jwtSecret string, reps ...repository.Repository) {
	admin := r.Group("/admin/import")
	admin.Use(middlewares.JWTAuthMiddleware(jwtSecret, reps...), middlewares.RequireActiveStatus(), middlewares.RequireRoleNames("admin"))
	{
		admin.POST("/excel/catalog", h.ImportCatalogExcel)
	}
}
