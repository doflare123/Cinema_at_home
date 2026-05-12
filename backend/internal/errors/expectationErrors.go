package errors

import "errors"

var (
	ErrInvalidExpectationTargetType = errors.New("invalid expectation target type")
	ErrInvalidExpectationVoteType   = errors.New("invalid expectation vote type")
	ErrInvalidExpectationTarget     = errors.New("target payload does not match target type")
	ErrInvalidExpectationScore      = errors.New("score must be between 1 and 10 when vote_type is score")
	ErrUnexpectedExpectationScore   = errors.New("score must be empty when vote_type is refuse")
)
