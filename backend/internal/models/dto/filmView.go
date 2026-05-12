package dto

type FilmView struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Poster      string `json:"poster"`
	ReleaseDate int    `json:"release_date"`
}
