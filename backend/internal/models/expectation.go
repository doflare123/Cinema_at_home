package models

import "time"

type ExpectationVote struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	TargetType  string    `gorm:"type:varchar(16);not null;index" json:"target_type"`
	MovieID     *uint     `gorm:"column:film_id;index" json:"movie_id,omitempty"`
	FranchiseID *uint     `gorm:"index" json:"franchise_id,omitempty"`
	VoteType    string    `gorm:"type:varchar(16);not null" json:"vote_type"`
	Score       *int      `json:"score,omitempty"`
	Comment     string    `gorm:"not null;default:''" json:"comment"`
	Movie       Film      `gorm:"foreignKey:MovieID" json:"movie,omitempty"`
	Franchise   Franchise `gorm:"foreignKey:FranchiseID" json:"franchise,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
