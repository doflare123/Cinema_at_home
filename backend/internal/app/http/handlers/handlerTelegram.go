package handlers

import (
	"cinema/internal/app/services"
	appErrors "cinema/internal/errors"
	"cinema/internal/models/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TelegramNotificationHandler interface {
	List(c *gin.Context)
	Enqueue(c *gin.Context)
	Retry(c *gin.Context)
}

type telegramNotificationHandler struct {
	service services.TelegramNotificationService
}

func NewTelegramNotificationHandler(service services.TelegramNotificationService) TelegramNotificationHandler {
	return &telegramNotificationHandler{service: service}
}

func (h *telegramNotificationHandler) List(c *gin.Context) {
	notifications, err := h.service.List(c.Query("status"), c.Query("type"))
	if err != nil {
		switch err {
		case appErrors.ErrInvalidTelegramNotificationStatus, appErrors.ErrInvalidTelegramNotificationType:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func (h *telegramNotificationHandler) Enqueue(c *gin.Context) {
	adminID, ok := currentUserID(c)
	if !ok {
		return
	}

	var req dto.EnqueueTelegramNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	notification, err := h.service.Enqueue(adminID, req)
	if err != nil {
		switch err {
		case appErrors.ErrInvalidTelegramNotificationType,
			appErrors.ErrInvalidTelegramNotificationPayload,
			appErrors.ErrInvalidTelegramNotificationEntity:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case appErrors.ErrWeeklyPackNotFound, appErrors.ErrUserNotFound, appErrors.ErrMovieProposalNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"notification": notification})
}

func (h *telegramNotificationHandler) Retry(c *gin.Context) {
	notificationID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	adminID, ok := currentUserID(c)
	if !ok {
		return
	}

	notification, err := h.service.Retry(notificationID, adminID)
	if err != nil {
		switch err {
		case appErrors.ErrTelegramNotificationNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case appErrors.ErrTelegramNotificationNotRetryable:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"notification": notification})
}
