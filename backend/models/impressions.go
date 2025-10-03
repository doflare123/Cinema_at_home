package models

type Impressions struct {
	ID       uint64 `json:"id"`
	UserID   uint64 `json:"user_id"`
	User     User   `gorm:"foreignKey:UserID"`
	FilmID   uint64 `json:"film_id"`
	Film     Films  `gorm:"foreignKey:FilmID"`
	Text     string `json:"text"`
	Score    uint32 `json:"score"`
	CreateAt string `json:"create_at"`
	UpdateAt string `json:"update_at"`
}
