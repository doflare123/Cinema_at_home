package dto

type CreateFranchiseRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type AddFranchiseMovieRequest struct {
	MovieID         uint `json:"movie_id" binding:"required"`
	PartNumber      int  `json:"part_number" binding:"required"`
	ReleaseOrder    int  `json:"release_order" binding:"required"`
	ChronologyOrder int  `json:"chronology_order" binding:"required"`
}

type FranchiseListItem struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	MoviesCount int    `json:"movies_count"`
}

type FranchiseMovieView struct {
	MovieID         uint   `json:"movie_id"`
	Title           string `json:"title"`
	Poster          string `json:"poster"`
	ReleaseDate     int    `json:"release_date"`
	PartNumber      int    `json:"part_number"`
	ReleaseOrder    int    `json:"release_order"`
	ChronologyOrder int    `json:"chronology_order"`
}

type FranchiseDetailView struct {
	ID          uint                 `json:"id"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Movies      []FranchiseMovieView `json:"movies"`
}
