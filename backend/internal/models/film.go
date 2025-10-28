package models

type Film struct {
	ID               uint    `gorm:"primaryKey" json:"id"`
	Title            string  `gorm:"not null" json:"title"`
	Description      string  `gorm:"not null" json:"description"`
	ShortDescription string  `gorm:"not null" json:"short_description"`
	Duration         string  `gorm:"not null" json:"duration"`
	ReleaseDate      string  `gorm:"not null" json:"release_date"`
	Country          string  `gorm:"not null" json:"country"`
	Poster           string  `gorm:"not null" json:"image"`
	RatinKp          float64 `gorm:"not null" json:"ratin"`
}

type FilmGenre struct {
	FilmID  uint
	GenreID uint
	Film    Film
	Genre   Genre
}

type Genre struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}
