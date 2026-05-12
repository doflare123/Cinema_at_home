package errors

import "errors"

var (
	ErrWeeklyPackNotFound           = errors.New("weekly pack not found")
	ErrInvalidWeeklyPackName        = errors.New("weekly pack name is required")
	ErrInvalidWeeklyPackSchedule    = errors.New("weekly pack ends_at must be after starts_at")
	ErrInvalidWeeklyPackStatus      = errors.New("invalid weekly pack status")
	ErrWeeklyPackStatusTransition   = errors.New("invalid weekly pack status transition")
	ErrWeeklyPackMustHaveMovies     = errors.New("weekly pack must contain at least one movie before voting starts")
	ErrWeeklyPackMovieAlreadyExists = errors.New("movie is already in weekly pack")
	ErrWeeklyPackMovieNotFound      = errors.New("movie is not part of weekly pack")
	ErrWeeklyPackMoviesLocked       = errors.New("weekly pack movies can only be edited in draft status")
	ErrInvalidWeeklyPackVoteScore   = errors.New("weekly pack vote must be one of: 3, 2, 1, 0, -2")
	ErrWeeklyPackVotingClosed       = errors.New("weekly pack is not open for voting")
	ErrWeeklyPackVoteLimitExceeded  = errors.New("weekly pack vote limit exceeded for this score")
)
