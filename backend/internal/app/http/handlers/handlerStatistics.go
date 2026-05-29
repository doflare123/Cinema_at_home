package handlers

import (
	"cinema/internal/app/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatisticsHandler interface {
	Summary(c *gin.Context)
}

type statisticsHandler struct {
	service services.StatisticsService
}

func NewStatisticsHandler(service services.StatisticsService) StatisticsHandler {
	return &statisticsHandler{service: service}
}

func (h *statisticsHandler) Summary(c *gin.Context) {
	summary, err := h.service.Summary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build statistics"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"summary": summary})
}
