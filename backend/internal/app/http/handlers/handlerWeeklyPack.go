package handlers

import (
	"cinema/internal/app/services"
	appErrors "cinema/internal/errors"
	"cinema/internal/models/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WeeklyPackHandler interface {
	List(c *gin.Context)
	GetByID(c *gin.Context)
	UpsertVote(c *gin.Context)
	MeVotes(c *gin.Context)
	Create(c *gin.Context)
	AddMovie(c *gin.Context)
	UpdateStatus(c *gin.Context)
}

type weeklyPackHandler struct {
	service services.WeeklyPackService
}

func NewWeeklyPackHandler(service services.WeeklyPackService) WeeklyPackHandler {
	return &weeklyPackHandler{service: service}
}

func (h *weeklyPackHandler) List(c *gin.Context) {
	items, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"weekly_packs": items})
}

func (h *weeklyPackHandler) GetByID(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	item, err := h.service.GetByID(id)
	if err != nil {
		if err == appErrors.ErrWeeklyPackNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"weekly_pack": item})
}

func (h *weeklyPackHandler) UpsertVote(c *gin.Context) {
	packID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	var req dto.UpsertWeeklyPackVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	vote, err := h.service.UpsertVote(packID, userID, req)
	if err != nil {
		switch err {
		case appErrors.ErrInvalidWeeklyPackVoteScore, appErrors.ErrWeeklyPackVoteLimitExceeded:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case appErrors.ErrWeeklyPackNotFound, appErrors.ErrWeeklyPackMovieNotFound, appErrors.ErrFilmNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case appErrors.ErrWeeklyPackVotingClosed:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"vote": vote})
}

func (h *weeklyPackHandler) MeVotes(c *gin.Context) {
	packID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	view, err := h.service.GetUserVotes(packID, userID)
	if err != nil {
		if err == appErrors.ErrWeeklyPackNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"votes": view})
}

func (h *weeklyPackHandler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req dto.CreateWeeklyPackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	item, err := h.service.Create(userID, req)
	if err != nil {
		switch err {
		case appErrors.ErrInvalidWeeklyPackName, appErrors.ErrInvalidWeeklyPackSchedule:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"weekly_pack": item})
}

func (h *weeklyPackHandler) AddMovie(c *gin.Context) {
	packID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.AddWeeklyPackMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	item, err := h.service.AddMovie(packID, req)
	if err != nil {
		switch err {
		case appErrors.ErrWeeklyPackMoviesLocked, appErrors.ErrWeeklyPackMovieAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case appErrors.ErrWeeklyPackNotFound, appErrors.ErrFilmNotFound, appErrors.ErrWeeklyPackMovieNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"weekly_pack": item})
}

func (h *weeklyPackHandler) UpdateStatus(c *gin.Context) {
	packID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateWeeklyPackStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	item, err := h.service.UpdateStatus(packID, req)
	if err != nil {
		switch err {
		case appErrors.ErrInvalidWeeklyPackStatus:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case appErrors.ErrWeeklyPackNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case appErrors.ErrWeeklyPackStatusTransition, appErrors.ErrWeeklyPackMustHaveMovies:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"weekly_pack": item})
}
