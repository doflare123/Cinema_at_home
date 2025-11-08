package handlers

import (
	"cinema/internal/app/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type filmReq struct {
	Title string `json:"title" binding:"required"`
}

type FilmHandler interface {
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
