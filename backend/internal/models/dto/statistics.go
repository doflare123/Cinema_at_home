package dto

import "time"

type StatisticsSummaryView struct {
	MoviesTotal           int64     `json:"movies_total"`
	FranchisesTotal       int64     `json:"franchises_total"`
	UsersActive           int64     `json:"users_active"`
	ProposalsPending      int64     `json:"proposals_pending"`
	ProposalsApproved     int64     `json:"proposals_approved"`
	ProposalsRejected     int64     `json:"proposals_rejected"`
	ExpectationsTotal     int64     `json:"expectations_total"`
	ExpectationsNumeric   int64     `json:"expectations_numeric"`
	ExpectationsRefuse    int64     `json:"expectations_refuse"`
	ExpectationsAverage   float64   `json:"expectations_average"`
	ReviewsTotal          int64     `json:"reviews_total"`
	ReviewsAverage        float64   `json:"reviews_average"`
	WeeklyPacksTotal      int64     `json:"weekly_packs_total"`
	WeeklyPacksVoting     int64     `json:"weekly_packs_voting"`
	WeeklyPacksClosed     int64     `json:"weekly_packs_closed"`
	WeeklyPackVotesTotal  int64     `json:"weekly_pack_votes_total"`
	WeeklyPackMoviesTotal int64     `json:"weekly_pack_movies_total"`
	GeneratedAt           time.Time `json:"generated_at"`
}
