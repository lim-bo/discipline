package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/limbo/discipline/pkg/entity"
)

type RegisterRequest struct {
	Name     string `validate:"required,alphanum_underscore,min=3,max=100"`
	Password string `validate:"required,min=8,max=72"`
}

type UserServiceI interface {
	// Validates user's credentials, creates new row in database. Returns user's data with ID.
	// If user with such name already exists, returns errorvalues.ErrUserExists
	Register(ctx context.Context, req *RegisterRequest) (*entity.User, error)
	// Compares given credentials to stored ones. If ok, give back user's data with ID.
	// If user not found, returns errorvalues.ErrUserNotFound.
	// If credentials are wrong, returns errorvalues.ErrWrongCredentials
	Login(ctx context.Context, name, password string) (*entity.User, error)
	// Searchs for user's metadata by given id.
	// If user not found, returns errorvalues.ErrUserNotFound
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	// Searchs for user's metadata by name.
	// If user not found, returns errorvalues.ErrUserNotFound
	GetByName(ctx context.Context, name string) (*entity.User, error)
	// Deletes user by id, needs password for security matters.
	// If user not found, returns errorvalues.ErrUserNotFound.
	// If password is wrong, returns errorvalues.ErrUserNotFound
	DeleteAccount(ctx context.Context, id uuid.UUID, password string) error
}

type CreateHabitRequest struct {
	Title       string
	Description string
}

type PaginationOpts struct {
	Limit  int
	Offset int
}

type HabitsServiceI interface {
	// Creates habit owned by user with uid. On success returns Habit data.
	// If there is no such owner (user), returns errorvalues.ErrUserNotFound
	CreateHabit(ctx context.Context, uid uuid.UUID, req CreateHabitRequest) (*entity.Habit, error)
	// Returns list of user's habits. Requires pagination options.
	// If there is no such user, returns empty list TO-DO: should check user for existion and return error, if doesn't exist
	GetUserHabits(ctx context.Context, uid uuid.UUID, pagination PaginationOpts) ([]*entity.Habit, error)
	// Deletes habit by habitID if userID is truly its owner.
	// If there is no habit with such ID, returns errorvalues.ErrHabitNotFound
	DeleteHabit(ctx context.Context, habitID, userID uuid.UUID) error
	// Returns habit metadata if userID is truly its owner.
	// If there is no habit with such ID, returns errorvalues.ErrHabitNotFound
	GetHabit(ctx context.Context, habitID, userID uuid.UUID) (*entity.Habit, error)
}

type HabitChecksServiceI interface {
	// Adds check to habit (habitID).
	// Compares userID with owner of habit with habitID, if they don't match, returns errovalues.ErrWrongOwner.
	// If there is attempt to create check to the future date, returns errorvalues.ErrCheckDateNotAllowed.
	// If there was check on this date already, returns errorvalues.ErrCheckExist
	CheckHabit(ctx context.Context, habitID, userID uuid.UUID, date time.Time) error
	// Unchecks habit (deletes check by date).
	// Compares userID with owner of habit with habitID, if they don't match, returns errovalues.ErrWrongOwner.
	// If there is no check on given date, returns errorvalues.ErrCheckNotFound
	UncheckHabit(ctx context.Context, habitID, userID uuid.UUID, date time.Time) error
	// Provides list of checks bound to given date interval.
	// Compares userID with owner of habit with habitID, if they don't match, returns errovalues.ErrWrongOwner.
	GetHabitChecks(ctx context.Context, habitID, userID uuid.UUID, from, to time.Time) ([]entity.HabitCheck, error)
	// Returns checks stat on habit.
	// Compares userID with owner of habit with habitID, if they don't match, returns errovalues.ErrWrongOwner.
	// Returns summ count of checks, streaks and last check date.
	GetHabitStats(ctx context.Context, habitID, userID uuid.UUID) (*entity.HabitStats, error)
}
