package errorvalues

import "errors"

var (
	ErrUserExists       = errors.New("such user already exists")
	ErrUserNotFound     = errors.New("user doesn't exists")
	ErrWrongCredentials = errors.New("wrong name or password")
)
