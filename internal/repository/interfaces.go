package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/limbo/discipline/pkg/entity"
)

type UsersRepositoryI interface {
	// Creates new user in database.
	// If user already exists, returns errorvalues.ErrUserExists
	Create(ctx context.Context, user *entity.User) error
	// Looks up user by name.
	// If there is no user with such name, returns errorvalues.ErrUserNotFound
	FindByName(ctx context.Context, name string) (*entity.User, error)
	// Looks up user by uid.
	// If there is no user with such uid, returns errorvalues.ErrUserNotFound
	FindByID(ctx context.Context, uid uuid.UUID) (*entity.User, error)
	// Updates user's info.
	// If there is no user with such uid to update, returns errorvalues.ErrUserNotFound
	Update(ctx context.Context, user *entity.User) error
	// Deletes user.
	// If there is no user with such uid to delete, returns errorvalues.ErrUserNotFound
	Delete(ctx context.Context, uid uuid.UUID) error
}

type HabitsRepositoryI interface {
	// Creates new habits in database. In habit only Title, UserID, Description are necessary.
	// If there was habit with such name and userID, returns errorvalues.ErrUserHasHabit.
	// If there is no user with owned habit, returns errorvalues.ErrOwnerNotFound
	Create(ctx context.Context, habit *entity.Habit) (uuid.UUID, error)
	// Searches habit with given id.
	// If there is not habit with such id, returns errorvalues.ErrHabitNotFound
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Habit, error)
	// Lists habits owned by user with uid. Requires pagination params provided.
	// If there is no habits owned by user or user doesn't exist, returns zero-len slice and nil.
	GetByUserID(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*entity.Habit, error)
	// Updates habit by ID (ID in habit is necessary).
	// If there is not habit with such id (in habit arg), returns errorvalues.ErrHabitNotFound
	Update(ctx context.Context, habit *entity.Habit) error
	// Deletes habit with id.
	// If there is not habit with such id, returns errorvalues.ErrHabitNotFound
	Delete(ctx context.Context, id uuid.UUID) error
}

type HabitChecksRepositoryI interface {
	// Creates new check on habit with habitID.
	// There is no habit for check, returns errorvalues.ErrHabitNotFound.
	// If habit was already checked, returns errorvalues.ErrCheckExist
	Create(ctx context.Context, habitID uuid.UUID, date time.Time) error
	// Deletes check on habit with habitID (uncheck).
	// If there is no such check, returns errorvalues.CheckNotFound
	Delete(ctx context.Context, habitID uuid.UUID, date time.Time) error
	// Inspects if check exists
	Exists(ctx context.Context, habitID uuid.UUID, date time.Time) (bool, error)
	// Provides checks of habitID for a period. If there is no habit with habitID,
	// returns zero-len slice and nil error.
	GetByHabitAndDateRange(ctx context.Context, habitID uuid.UUID, from, to time.Time) ([]entity.HabitCheck, error)
	// Returns date of last check on habitID. If there is no checks on habit,
	// returns nil time and nil error.
	GetLastCheckDate(ctx context.Context, habitID uuid.UUID) (*time.Time, error)
	// Returns count of checks for habitID. If there is no habit with habitID,
	// returns 0 and nil error.
	CountByHabitID(ctx context.Context, habitID uuid.UUID) (int, error)
}

type DBConfig interface {
	ConnString() string
}

type PgConnection interface {
	Ping(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PGCfg struct {
	Address  string
	Username string
	Password string
	DB       string
}

func (pgcfg *PGCfg) ConnString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s/%s", pgcfg.Username, pgcfg.Password, pgcfg.Address, pgcfg.DB)
}
