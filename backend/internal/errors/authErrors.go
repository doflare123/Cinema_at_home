package errors

import "errors"

var (
	ErrUserNameAlreadyExist = errors.New("username already exist")
	ErrInvalidPassword      = errors.New("invalid password or username")
	ErrInvalidRefreshToken  = errors.New("invalid refresh token")
	ErrUpdDataUser          = errors.New("Couldn't update user data")
	ErrEmptyPassword        = errors.New("empty password")
	ErrInvalidServer        = errors.New("invalid server")
	ErrUserNotFound         = errors.New("user not found")
	ErrUserNotActive        = errors.New("user is not active yet")
	ErrInvalidTelegramAuth  = errors.New("invalid telegram auth payload")
	ErrProblemWithCreateJWT = errors.New("problem with create jwt")
)
