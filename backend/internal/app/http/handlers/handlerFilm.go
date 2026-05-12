package handlers

import (
	"cinema/internal/app/services"
	appErrors "cinema/internal/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type filmReq struct {
	Title string `json:"title" binding:"required"`
}

type FilmHandler interface {
	List(c *gin.Context)
	GetByID(c *gin.Context)
	CreateFilm(c *gin.Context)
}

type filmHandler struct {
	srv services.FilmService
}

func NewFilmHandler(srv services.FilmService) FilmHandler {
	return &filmHandler{
		srv: srv,
	}
}

func (h *filmHandler) List(c *gin.Context) {
	items, err := h.srv.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movies": items})
}

func (h *filmHandler) GetByID(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	item, err := h.srv.GetByID(id)
	if err != nil {
		if err == appErrors.ErrFilmNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movie": item})
}

func (h *filmHandler) CreateFilm(c *gin.Context) {
	var req filmReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.srv.Create(req.Title); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Film added successfully"})
}
