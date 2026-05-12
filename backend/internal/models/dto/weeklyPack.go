package dto

import "time"

type CreateWeeklyPackRequest struct {
	Name     string     `json:"name" binding:"required"`
	StartsAt *time.Time `json:"starts_at,omitempty"`
	EndsAt   *time.Time `json:"ends_at,omitempty"`
}

type AddWeeklyPackMovieRequest struct {
	MovieID   uint `json:"movie_id" binding:"required"`
	SortOrder int  `json:"sort_order"`
}

type UpdateWeeklyPackStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type UpsertWeeklyPackVoteRequest struct {
	MovieID uint `json:"movie_id" binding:"required"`
	Score   *int `json:"score"`
}

type WeeklyPackListItem struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Status          string     `json:"status"`
	StartsAt        *time.Time `json:"starts_at,omitempty"`
	EndsAt          *time.Time `json:"ends_at,omitempty"`
	CreatedByUserID uint       `json:"created_by_user_id"`
	MoviesCount     int        `json:"movies_count"`
	VotesCount      int        `json:"votes_count"`
}

type WeeklyPackVoteBreakdownItem struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Score       int    `json:"score"`
}

type WeeklyPackMovieView struct {
	MovieID      uint                         `json:"movie_id"`
	Title        string                       `json:"title"`
	Poster       string                       `json:"poster"`
	ReleaseDate  int                          `json:"release_date"`
	SortOrder    int                          `json:"sort_order"`
	ScoreTotal   int                          `json:"score_total"`
	Plus3Count   int                          `json:"plus_3_count"`
	Plus2Count   int                          `json:"plus_2_count"`
	Plus1Count   int                          `json:"plus_1_count"`
	ZeroCount    int                          `json:"zero_count"`
	Minus2Count  int                          `json:"minus_2_count"`
	Votes        []WeeklyPackVoteBreakdownItem `json:"votes"`
}

type WeeklyPackDetailView struct {
	ID              uint                `json:"id"`
	Name            string              `json:"name"`
	Status          string              `json:"status"`
	StartsAt        *time.Time          `json:"starts_at,omitempty"`
	EndsAt          *time.Time          `json:"ends_at,omitempty"`
	CreatedByUserID uint                `json:"created_by_user_id"`
	Movies          []WeeklyPackMovieView `json:"movies"`
}

type WeeklyPackUserVoteItem struct {
	MovieID uint `json:"movie_id"`
	Score   int  `json:"score"`
}

type WeeklyPackUserVotesView struct {
	PackID uint                    `json:"pack_id"`
	Votes  []WeeklyPackUserVoteItem `json:"votes"`
}
