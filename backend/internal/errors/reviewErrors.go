package errors

import "errors"

var (
	ErrReviewNotFound             = errors.New("review not found")
	ErrInvalidReviewMode          = errors.New("invalid review mode")
	ErrInvalidReviewScore         = errors.New("simple mode score must be between 1 and 10")
	ErrReviewScoreRequired        = errors.New("score is required for simple mode")
	ErrUnexpectedReviewScore      = errors.New("score must be empty for criteria mode")
	ErrReviewCriteriaRequired     = errors.New("criteria_scores are required for criteria mode")
	ErrUnexpectedReviewCriteria   = errors.New("criteria_scores must be empty for simple mode")
	ErrInvalidReviewCriterion     = errors.New("criteria mode scores must be between 1 and 10")
	ErrManualReviewFinalScoreEdit = errors.New("final_score is calculated automatically and cannot be provided")
)
