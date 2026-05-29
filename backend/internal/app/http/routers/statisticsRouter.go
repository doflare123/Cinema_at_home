package routers

import (
	"cinema/internal/app/http/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterStatisticsRoutes(r *gin.Engine, h handlers.StatisticsHandler) {
	r.GET("/statistics/summary", h.Summary)
}
