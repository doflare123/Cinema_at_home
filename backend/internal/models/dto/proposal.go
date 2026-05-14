package dto

import "time"

type CreateMovieProposalRequest struct {
	Title            string  `json:"title" binding:"required"`
	Description      string  `json:"description" binding:"required"`
	SmallDescription string  `json:"small_description" binding:"required"`
	Duration         int32   `json:"duration" binding:"required"`
	ReleaseDate      int     `json:"release_date" binding:"required"`
	Country          string  `json:"country" binding:"required"`
	Poster           string  `json:"poster" binding:"required"`
	RatingKp         float64 `json:"rating_kp"`
	Source           string  `json:"source"`
}

type ModerateMovieProposalRequest struct {
	Status            string `json:"status" binding:"required"`
	ModerationComment string `json:"moderation_comment"`
}

type MovieProposalView struct {
	ID                  uint       `json:"id"`
	Title               string     `json:"title"`
	Description         string     `json:"description"`
	SmallDescription    string     `json:"small_description"`
	Duration            int32      `json:"duration"`
	ReleaseDate         int        `json:"release_date"`
	Country             string     `json:"country"`
	Poster              string     `json:"poster"`
	RatingKp            float64    `json:"rating_kp"`
	Source              string     `json:"source"`
	Status              string     `json:"status"`
	ProposedByUserID    uint       `json:"proposed_by_user_id"`
	ProposedByUsername  string     `json:"proposed_by_username"`
	ProposedByName      string     `json:"proposed_by_name"`
	ModeratedByUserID   *uint      `json:"moderated_by_user_id,omitempty"`
	ModeratedByUsername string     `json:"moderated_by_username,omitempty"`
	ModeratedByName     string     `json:"moderated_by_name,omitempty"`
	ModeratedAt         *time.Time `json:"moderated_at,omitempty"`
	ModerationComment   string     `json:"moderation_comment"`
	FilmID              *uint      `json:"film_id,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}
