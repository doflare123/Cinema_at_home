package models

import "time"

type WeeklyPack struct {
	ID              uint              `gorm:"primaryKey" json:"id"`
	Name            string            `gorm:"not null" json:"name"`
	Status          string            `gorm:"type:varchar(16);not null;index" json:"status"`
	StartsAt        *time.Time        `json:"starts_at,omitempty"`
	EndsAt          *time.Time        `json:"ends_at,omitempty"`
	CreatedByUserID uint              `gorm:"not null;index" json:"created_by_user_id"`
	CreatedByUser   User              `gorm:"foreignKey:CreatedByUserID" json:"created_by_user,omitempty"`
	Movies          []WeeklyPackMovie `gorm:"foreignKey:PackID" json:"movies,omitempty"`
	Votes           []WeeklyPackVote  `gorm:"foreignKey:PackID" json:"votes,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type WeeklyPackMovie struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PackID    uint      `gorm:"not null;uniqueIndex:idx_weekly_pack_movie" json:"pack_id"`
	MovieID   uint      `gorm:"column:film_id;not null;uniqueIndex:idx_weekly_pack_movie;index" json:"movie_id"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	Movie     Film      `gorm:"foreignKey:MovieID" json:"movie"`
	CreatedAt time.Time `json:"created_at"`
}

type WeeklyPackVote struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PackID    uint      `gorm:"not null;uniqueIndex:idx_weekly_pack_vote;index" json:"pack_id"`
	MovieID   uint      `gorm:"column:film_id;not null;uniqueIndex:idx_weekly_pack_vote;index" json:"movie_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_weekly_pack_vote;index" json:"user_id"`
	Score     int       `gorm:"column:vote_value;not null" json:"score"`
	LimitSlot *int      `gorm:"column:limit_slot" json:"-"`
	Movie     Film      `gorm:"foreignKey:MovieID" json:"movie"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
