package handlers

import (
	"cinema/internal/app/services"
	appErrors "cinema/internal/errors"
	"cinema/internal/models/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReviewHandler interface {
	ListByFilm(c *gin.Context)
	Me(c *gin.Context)
	Upsert(c *gin.Context)
}

type reviewHandler struct {
	service services.ReviewService
}

func NewReviewHandler(service services.ReviewService) ReviewHandler {
	return &reviewHandler{service: service}
}

func (h *reviewHandler) ListByFilm(c *gin.Context) {
	filmID, ok := parseUintParam(c, "filmId")
	if !ok {
		return
	}

	reviews, err := h.service.ListByFilm(filmID)
	if err != nil {
		if err == appErrors.ErrFilmNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}

func (h *reviewHandler) Me(c *gin.Context) {
	filmID, ok := parseUintParam(c, "filmId")
	if !ok {
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	review, err := h.service.GetByFilmAndUser(filmID, userID)
	if err != nil {
		switch err {
		case appErrors.ErrFilmNotFound, appErrors.ErrReviewNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"review": review})
}

func (h *reviewHandler) Upsert(c *gin.Context) {
	filmID, ok := parseUintParam(c, "filmId")
	if !ok {
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	var req dto.UpsertReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	review, err := h.service.Upsert(filmID, userID, req)
	if err != nil {
		switch err {
		case appErrors.ErrInvalidReviewMode,
			appErrors.ErrInvalidReviewScore,
			appErrors.ErrReviewScoreRequired,
			appErrors.ErrUnexpectedReviewScore,
			appErrors.ErrReviewCriteriaRequired,
			appErrors.ErrUnexpectedReviewCriteria,
			appErrors.ErrInvalidReviewCriterion,
			appErrors.ErrManualReviewFinalScoreEdit:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case appErrors.ErrFilmNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"review": review})
}
