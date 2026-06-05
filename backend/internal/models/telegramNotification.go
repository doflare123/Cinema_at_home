package models

import "time"

type TelegramNotification struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Type             string     `gorm:"type:varchar(32);not null;index" json:"type"`
	Status           string     `gorm:"type:varchar(16);not null;index" json:"status"`
	TargetUserID     *uint      `gorm:"column:target_user_id;index" json:"target_user_id,omitempty"`
	TargetUser       User       `gorm:"foreignKey:TargetUserID" json:"target_user,omitempty"`
	EntityID         uint       `gorm:"column:entity_id;not null;index" json:"entity_id"`
	Payload          string     `gorm:"type:jsonb;not null" json:"payload"`
	Attempts         int        `gorm:"not null;default:0" json:"attempts"`
	LastError        string     `gorm:"column:last_error;not null;default:''" json:"last_error"`
	EnqueuedByUserID uint       `gorm:"column:enqueued_by_user_id;not null;index" json:"enqueued_by_user_id"`
	EnqueuedBy       User       `gorm:"foreignKey:EnqueuedByUserID" json:"enqueued_by,omitempty"`
	SentAt           *time.Time `gorm:"column:sent_at" json:"sent_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
