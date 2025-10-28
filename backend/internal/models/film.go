package models

type Film struct {
	ID               uint    `gorm:"primaryKey"`
	Title            string  `gorm:"not null"`
	Description      string  `gorm:"not null"`
	ShortDescription string  `gorm:"not null"`
	Duration         int32   `gorm:"not null"`
	ReleaseDate      string  `gorm:"not null"`
	Country          string  `gorm:"not null"`
	Poster           string  `gorm:"not null"`
	RatinKp          float64 `gorm:"not null"`
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
