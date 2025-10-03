package models

type Films struct {
	ID          uint64     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Rating      uint32     `json:"rating"`
	Image       string     `json:"image"`
	StatusID    uint64     `json:"status_id"`
	Status      FilmStatus `gorm:"foreignKey:StatusID"`
	CreateAt    string     `json:"create_at"`
	UpdateAt    string     `json:"update_at"`
}
