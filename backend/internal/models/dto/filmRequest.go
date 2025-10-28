package dto

type AboutFilm struct {
	Title            string   `json:"title" binding:"required"`
	Description      string   `json:"description" binding:"required"`
	ShortDescription string   `json:"short_description" binding:"required"`
	Duration         int      `json:"duration" binding:"required"`
	ReleaseDate      string   `json:"release_date" binding:"required"`
	Country          string   `json:"country" binding:"required"`
	Poster           string   `json:"poster" binding:"required"`
	RatingKp         int      `json:"rating" binding:"required,min=0,max=10"`
	Genre            []string `json:"genre" binding:"required"`
}

type Film struct {
	ID               uint     `json:"id"`
	Title            string   `json:"title"`
	ShortDescription string   `json:"short_description"`
	Duration         int32    `json:"duration"`
	Poster           string   `json:"poster"`
	RatinKp          float64  `json:"rating"`
	Genre            []string `json:"genre"`
}
