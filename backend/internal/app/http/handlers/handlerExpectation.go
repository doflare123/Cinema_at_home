package handlers

import (
	"cinema/internal/app/services"
	appErrors "cinema/internal/errors"
	"cinema/internal/models/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ExpectationHandler interface {
	Upsert(c *gin.Context)
	Summary(c *gin.Context)
	Me(c *gin.Context)
}

type expectationHandler struct {
	service services.ExpectationService
}

func NewExpectationHandler(service services.ExpectationService) ExpectationHandler {
	return &expectationHandler{service: service}
}

func (h *expectationHandler) Upsert(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	var request dto.UpsertExpectationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	vote, err := h.service.Upsert(userID, request)
	if err != nil {
		switch err {
		case appErrors.ErrInvalidExpectationTargetType,
			appErrors.ErrInvalidExpectationTarget,
			appErrors.ErrInvalidExpectationVoteType,
			appErrors.ErrInvalidExpectationScore,
			appErrors.ErrUnexpectedExpectationScore:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case appErrors.ErrFilmNotFound, appErrors.ErrFranchiseNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"vote": vote})
}

func (h *expectationHandler) Summary(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	summary, err := h.service.Summary(c.Param("targetType"), id)
	if err != nil {
		switch err {
		case appErrors.ErrInvalidExpectationTargetType:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case appErrors.ErrFilmNotFound, appErrors.ErrFranchiseNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

func (h *expectationHandler) Me(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	votes, err := h.service.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"votes": votes})
}

func currentUserID(c *gin.Context) (uint, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return 0, false
	}

	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return 0, false
	}

	return userID, true
}
