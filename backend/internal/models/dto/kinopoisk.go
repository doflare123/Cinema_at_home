package dto

type KinopoiskSearchResult struct {
	ProviderMovieID  string   `json:"provider_movie_id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	SmallDescription string   `json:"small_description"`
	Duration         int32    `json:"duration"`
	ReleaseDate      int      `json:"release_date"`
	Country          string   `json:"country"`
	Poster           string   `json:"poster"`
	RatingKp         float64  `json:"rating_kp"`
	Genres           []string `json:"genres"`
	Source           string   `json:"source"`
}
