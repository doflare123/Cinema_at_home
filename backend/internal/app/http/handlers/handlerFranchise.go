package handlers

import (
	"cinema/internal/app/services"
	appErrors "cinema/internal/errors"
	"cinema/internal/models/dto"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FranchiseHandler interface {
	List(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	AddMovie(c *gin.Context)
}

type franchiseHandler struct {
	service services.FranchiseService
}

func NewFranchiseHandler(service services.FranchiseService) FranchiseHandler {
	return &franchiseHandler{service: service}
}

func (h *franchiseHandler) List(c *gin.Context) {
	items, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"franchises": items})
}

func (h *franchiseHandler) GetByID(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	franchise, err := h.service.GetByID(id)
	if err != nil {
		if err == appErrors.ErrFranchiseNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"franchise": franchise})
}

func (h *franchiseHandler) Create(c *gin.Context) {
	var request dto.CreateFranchiseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	franchise, err := h.service.Create(request)
	if err != nil {
		switch err {
		case appErrors.ErrInvalidFranchiseTitle:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case appErrors.ErrFranchiseTitleAlreadyExist:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"franchise": franchise})
}

func (h *franchiseHandler) AddMovie(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var request dto.AddFranchiseMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	franchise, err := h.service.AddMovie(id, request)
	if err != nil {
		switch err {
		case appErrors.ErrFranchiseNotFound, appErrors.ErrFilmNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case appErrors.ErrMovieAlreadyInFranchise:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case appErrors.ErrInvalidFranchiseMovieLink:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"franchise": franchise})
}

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return uint(value), true
}
