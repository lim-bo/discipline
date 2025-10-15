package errorvalues

import "errors"

var (
	ErrUserExists          = errors.New("such user already exists")
	ErrUserNotFound        = errors.New("user doesn't exists")
	ErrWrongCredentials    = errors.New("wrong name or password")
	ErrInvalidToken        = errors.New("invalid token")
	ErrUserHasHabit        = errors.New("habit with such title already owned by user")
	ErrHabitNotFound       = errors.New("habit doesn't exists")
	ErrOwnerNotFound       = errors.New("user to own habit not found")
	ErrWrongOwner          = errors.New("habit owner and given user don't match")
	ErrCheckExist          = errors.New("habit already checked on this date")
	ErrCheckNotFound       = errors.New("habit check on this date not found")
	ErrCheckDateNotAllowed = errors.New("can't check habit on date in the future")
)
