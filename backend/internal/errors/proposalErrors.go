package errors

import "errors"

var (
	ErrMovieProposalNotFound       = errors.New("movie proposal not found")
	ErrInvalidMovieProposalStatus  = errors.New("invalid movie proposal status")
	ErrMovieProposalAlreadyClosed  = errors.New("movie proposal is already moderated")
	ErrInvalidMovieProposalPayload = errors.New("movie proposal payload is invalid")
	ErrMovieProposalDuplicateFilm  = errors.New("movie proposal cannot be approved because movie already exists")
)
