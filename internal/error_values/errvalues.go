package errorvalues

import "errors"

var (
	ErrUserExists       = errors.New("such user already exists")
	ErrUserNotFound     = errors.New("user doesn't exists")
	ErrWrongCredentials = errors.New("wrong name or password")
	ErrInvalidToken     = errors.New("invalid token")
	ErrUserHasHabit     = errors.New("habit with such title already owned by user")
	ErrHabitNotFound    = errors.New("habit doesn't exists")
)
