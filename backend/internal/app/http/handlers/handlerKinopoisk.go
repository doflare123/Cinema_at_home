package handlers

import (
	"cinema/internal/app/services"
	appErrors "cinema/internal/errors"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type KinopoiskHandler interface {
	Search(c *gin.Context)
}

type kinopoiskHandler struct {
	service services.KinopoiskService
}

func NewKinopoiskHandler(service services.KinopoiskService) KinopoiskHandler {
	return &kinopoiskHandler{service: service}
}

func (h *kinopoiskHandler) Search(c *gin.Context) {
	limit := 10
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
		limit = parsed
	}

	query := c.Query("q")
	if query == "" {
		query = c.Query("query")
	}

	results, err := h.service.Search(query, limit)
	if err != nil {
		switch {
		case errors.Is(err, appErrors.ErrKinopoiskQueryRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, appErrors.ErrKinopoiskSearchFailed):
			c.JSON(http.StatusBadGateway, gin.H{"error": appErrors.ErrKinopoiskSearchFailed.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}
