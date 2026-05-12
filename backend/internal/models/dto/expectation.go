package dto

type UpsertExpectationRequest struct {
	TargetType  string `json:"target_type" binding:"required"`
	MovieID     *uint  `json:"movie_id,omitempty"`
	FranchiseID *uint  `json:"franchise_id,omitempty"`
	VoteType    string `json:"vote_type" binding:"required"`
	Score       *int   `json:"score,omitempty"`
	Comment     string `json:"comment"`
}

type ExpectationVoteView struct {
	ID          uint   `json:"id"`
	TargetType  string `json:"target_type"`
	MovieID     *uint  `json:"movie_id,omitempty"`
	FranchiseID *uint  `json:"franchise_id,omitempty"`
	TargetTitle string `json:"target_title"`
	VoteType    string `json:"vote_type"`
	Score       *int   `json:"score,omitempty"`
	Comment     string `json:"comment"`
}

type ExpectationSummaryView struct {
	TargetType      string  `json:"target_type"`
	TargetID        uint    `json:"target_id"`
	Avg             float64 `json:"avg"`
	NumericCount    int64   `json:"numeric_count"`
	RefuseCount     int64   `json:"refuse_count"`
	ThresholdPassed bool    `json:"threshold_passed"`
}
