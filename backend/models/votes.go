package models

type Votes struct {
	ID     uint64 `json:"id"`
	UserID uint64 `json:"user_id"`
	User   User   `gorm:"foreignKey:UserID"`
	FilmID uint64 `json:"film_id"`
	Film   Films  `gorm:"foreignKey:FilmID"`
	Vote   uint32 `json:"vote"  gorm:"not null"`
}
