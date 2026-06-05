package dto

import (
	"encoding/json"
	"time"
)

type EnqueueTelegramNotificationRequest struct {
	Type         string          `json:"type" binding:"required"`
	EntityID     uint            `json:"entity_id" binding:"required"`
	TargetUserID *uint           `json:"target_user_id,omitempty"`
	Payload      json.RawMessage `json:"payload,omitempty"`
}

type RetryTelegramNotificationRequest struct {
	LastError string `json:"last_error,omitempty"`
}

type TelegramNotificationView struct {
	ID                 uint            `json:"id"`
	Type               string          `json:"type"`
	Status             string          `json:"status"`
	TargetUserID       *uint           `json:"target_user_id,omitempty"`
	TargetUsername     string          `json:"target_username,omitempty"`
	TargetDisplayName  string          `json:"target_display_name,omitempty"`
	EntityID           uint            `json:"entity_id"`
	Payload            json.RawMessage `json:"payload"`
	Attempts           int             `json:"attempts"`
	LastError          string          `json:"last_error,omitempty"`
	EnqueuedByUserID   uint            `json:"enqueued_by_user_id"`
	EnqueuedByUsername string          `json:"enqueued_by_username,omitempty"`
	EnqueuedByName     string          `json:"enqueued_by_name,omitempty"`
	SentAt             *time.Time      `json:"sent_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}
