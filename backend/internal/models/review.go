package models

import "time"

type Review struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	FilmID             uint      `gorm:"column:film_id;not null;uniqueIndex:idx_review_film_user;index" json:"film_id"`
	UserID             uint      `gorm:"not null;uniqueIndex:idx_review_film_user;index" json:"user_id"`
	Mode               string    `gorm:"type:varchar(16);not null" json:"mode"`
	Score              *int      `json:"score,omitempty"`
	FinalScore         float64   `gorm:"column:final_score;not null" json:"final_score"`
	CriteriaScoresJSON string    `gorm:"column:criteria_scores;type:jsonb;not null;default:'{}'" json:"-"`
	Comment            string    `gorm:"not null;default:''" json:"comment"`
	Film               Film      `gorm:"foreignKey:FilmID" json:"film,omitempty"`
	User               User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
