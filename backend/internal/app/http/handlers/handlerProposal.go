package handlers

import (
	"cinema/internal/app/services"
	appErrors "cinema/internal/errors"
	"cinema/internal/models/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MovieProposalHandler interface {
	Create(c *gin.Context)
	My(c *gin.Context)
	List(c *gin.Context)
	Moderate(c *gin.Context)
}

type movieProposalHandler struct {
	service services.MovieProposalService
}

func NewMovieProposalHandler(service services.MovieProposalService) MovieProposalHandler {
	return &movieProposalHandler{service: service}
}

func (h *movieProposalHandler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	var req dto.CreateMovieProposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	proposal, err := h.service.Create(userID, req)
	if err != nil {
		if err == appErrors.ErrInvalidMovieProposalPayload {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"proposal": proposal})
}

func (h *movieProposalHandler) My(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	proposals, err := h.service.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"proposals": proposals})
}

func (h *movieProposalHandler) List(c *gin.Context) {
	proposals, err := h.service.List(c.Query("status"))
	if err != nil {
		if err == appErrors.ErrInvalidMovieProposalStatus {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"proposals": proposals})
}

func (h *movieProposalHandler) Moderate(c *gin.Context) {
	proposalID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	adminID, ok := currentUserID(c)
	if !ok {
		return
	}

	var req dto.ModerateMovieProposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	proposal, err := h.service.Moderate(proposalID, adminID, req)
	if err != nil {
		switch err {
		case appErrors.ErrMovieProposalNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case appErrors.ErrInvalidMovieProposalStatus,
			appErrors.ErrMovieProposalAlreadyClosed,
			appErrors.ErrMovieProposalDuplicateFilm:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"proposal": proposal})
}
