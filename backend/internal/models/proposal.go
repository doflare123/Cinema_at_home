package models

import "time"

type MovieProposal struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Title             string     `gorm:"not null;index" json:"title"`
	Description       string     `gorm:"not null" json:"description"`
	SmallDescription  string     `gorm:"column:small_description;not null" json:"small_description"`
	Duration          int32      `gorm:"not null" json:"duration"`
	ReleaseDate       int        `gorm:"column:release_date;not null" json:"release_date"`
	Country           string     `gorm:"not null" json:"country"`
	Poster            string     `gorm:"not null" json:"poster"`
	RatingKp          float64    `gorm:"column:rating_kp;not null" json:"rating_kp"`
	Source            string     `gorm:"type:varchar(32);not null;default:'manual';index" json:"source"`
	Status            string     `gorm:"type:varchar(16);not null;default:'pending';index" json:"status"`
	ProposedByUserID  uint       `gorm:"column:proposed_by_user_id;not null;index" json:"proposed_by_user_id"`
	ModeratedByUserID *uint      `gorm:"column:moderated_by_user_id;index" json:"moderated_by_user_id,omitempty"`
	ModeratedAt       *time.Time `gorm:"column:moderated_at" json:"moderated_at,omitempty"`
	ModerationComment string     `gorm:"column:moderation_comment;not null;default:''" json:"moderation_comment"`
	FilmID            *uint      `gorm:"column:film_id;index" json:"film_id,omitempty"`
	ProposedBy        User       `gorm:"foreignKey:ProposedByUserID" json:"proposed_by,omitempty"`
	ModeratedBy       User       `gorm:"foreignKey:ModeratedByUserID" json:"moderated_by,omitempty"`
	Film              Film       `gorm:"foreignKey:FilmID" json:"film,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
