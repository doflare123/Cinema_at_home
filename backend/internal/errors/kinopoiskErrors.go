package errors

import "errors"

var (
	ErrKinopoiskQueryRequired = errors.New("kinopoisk query is required")
	ErrKinopoiskSearchFailed  = errors.New("kinopoisk search failed")
	ErrKinopoiskNotConfigured = errors.New("kinopoisk integration is not configured")
)
