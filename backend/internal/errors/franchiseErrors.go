package errors

import "errors"

var (
	ErrFranchiseNotFound          = errors.New("franchise not found")
	ErrInvalidFranchiseTitle      = errors.New("franchise title is required")
	ErrFranchiseTitleAlreadyExist = errors.New("franchise title already exists")
	ErrMovieAlreadyInFranchise    = errors.New("movie already linked to franchise")
	ErrInvalidFranchiseMovieLink  = errors.New("movie_id, part_number, release_order and chronology_order must be positive")
)
