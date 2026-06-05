package handlers

import (
	"cinema/internal/models/dto"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appErrors "cinema/internal/errors"

	"github.com/gin-gonic/gin"
)

type stubTelegramNotificationService struct {
	list    func(status, notificationType string) ([]dto.TelegramNotificationView, error)
	enqueue func(enqueuedByUserID uint, req dto.EnqueueTelegramNotificationRequest) (dto.TelegramNotificationView, error)
	retry   func(notificationID, adminID uint) (dto.TelegramNotificationView, error)
}

func (s stubTelegramNotificationService) List(status, notificationType string) ([]dto.TelegramNotificationView, error) {
	if s.list != nil {
		return s.list(status, notificationType)
	}
	return nil, nil
}

func (s stubTelegramNotificationService) Enqueue(enqueuedByUserID uint, req dto.EnqueueTelegramNotificationRequest) (dto.TelegramNotificationView, error) {
	if s.enqueue != nil {
		return s.enqueue(enqueuedByUserID, req)
	}
	return dto.TelegramNotificationView{}, nil
}

func (s stubTelegramNotificationService) Retry(notificationID, adminID uint) (dto.TelegramNotificationView, error) {
	if s.retry != nil {
		return s.retry(notificationID, adminID)
	}
	return dto.TelegramNotificationView{}, nil
}

func TestTelegramNotificationHandlerEnqueueReturnsCreated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewTelegramNotificationHandler(stubTelegramNotificationService{
		enqueue: func(enqueuedByUserID uint, req dto.EnqueueTelegramNotificationRequest) (dto.TelegramNotificationView, error) {
			return dto.TelegramNotificationView{
				ID:               10,
				Type:             req.Type,
				Status:           "queued",
				EntityID:         req.EntityID,
				EnqueuedByUserID: enqueuedByUserID,
			}, nil
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("user_id", uint(7))
	ctx.Request = httptest.NewRequest(http.MethodPost, "/admin/telegram/notifications", strings.NewReader(`{
		"type":"weekly_pack_results",
		"entity_id":42,
		"payload":{"message":"ready"}
	}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Enqueue(ctx)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
}

func TestTelegramNotificationHandlerEnqueueMapsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewTelegramNotificationHandler(stubTelegramNotificationService{
		enqueue: func(enqueuedByUserID uint, req dto.EnqueueTelegramNotificationRequest) (dto.TelegramNotificationView, error) {
			return dto.TelegramNotificationView{}, appErrors.ErrWeeklyPackNotFound
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("user_id", uint(7))
	ctx.Request = httptest.NewRequest(http.MethodPost, "/admin/telegram/notifications", strings.NewReader(`{"type":"weekly_pack_results","entity_id":999}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Enqueue(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestTelegramNotificationHandlerRetryMapsNotRetryable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewTelegramNotificationHandler(stubTelegramNotificationService{
		retry: func(notificationID, adminID uint) (dto.TelegramNotificationView, error) {
			return dto.TelegramNotificationView{}, appErrors.ErrTelegramNotificationNotRetryable
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: "42"}}
	ctx.Set("user_id", uint(7))
	ctx.Request = httptest.NewRequest(http.MethodPost, "/admin/telegram/notifications/42/retry", nil)

	handler.Retry(ctx)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", recorder.Code)
	}
}
