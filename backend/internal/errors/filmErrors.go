package errors

import "errors"

var (
	ErrFilmNotFound = errors.New("Film is not found in db")
)
