package models

import "time"

type Franchise struct {
	ID          uint             `gorm:"primaryKey" json:"id"`
	Title       string           `gorm:"not null;uniqueIndex" json:"title"`
	Description string           `gorm:"not null;default:''" json:"description"`
	Movies      []FranchiseMovie `gorm:"foreignKey:FranchiseID" json:"movies,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type FranchiseMovie struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	FranchiseID     uint      `gorm:"not null;uniqueIndex:idx_franchise_movie" json:"franchise_id"`
	MovieID         uint      `gorm:"column:film_id;not null;uniqueIndex:idx_franchise_movie" json:"movie_id"`
	PartNumber      int       `gorm:"not null" json:"part_number"`
	ReleaseOrder    int       `gorm:"not null" json:"release_order"`
	ChronologyOrder int       `gorm:"not null" json:"chronology_order"`
	Movie           Film      `gorm:"foreignKey:MovieID" json:"movie"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
