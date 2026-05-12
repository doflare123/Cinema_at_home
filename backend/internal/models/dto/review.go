package dto

import "time"

type UpsertReviewRequest struct {
	Mode           string         `json:"mode" binding:"required"`
	Score          *int           `json:"score,omitempty"`
	CriteriaScores map[string]int `json:"criteria_scores,omitempty"`
	FinalScore     *float64       `json:"final_score,omitempty"`
	Comment        string         `json:"comment"`
}

type ReviewView struct {
	ID             uint           `json:"id"`
	FilmID         uint           `json:"film_id"`
	UserID         uint           `json:"user_id"`
	Username       string         `json:"username"`
	DisplayName    string         `json:"display_name"`
	Mode           string         `json:"mode"`
	Score          *int           `json:"score,omitempty"`
	CriteriaScores map[string]int `json:"criteria_scores,omitempty"`
	FinalScore     float64        `json:"final_score"`
	Comment        string         `json:"comment"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}
