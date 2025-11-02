package errors

import "errors"

var (
	ErrUserNameAlreadyExist = errors.New("username already exist")
	ErrInvalidPassword      = errors.New("invalid password or username")
	ErrUpdDataUser          = errors.New("Couldn't update user data")
	ErrEmptyPassword        = errors.New("empty password")
	ErrInvalidServer        = errors.New("invalid server")
	ErrUserNotFound         = errors.New("user not found")
	ErrProblemWithCreateJWT = errors.New("problem with create jwt")
)
