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
	// Creates new user in database
	Create(ctx context.Context, user *entity.User) error
	// Looks up user by name. Can be used for login
	FindByName(ctx context.Context, name string) (*entity.User, error)
	// Looks up user by uid. Can be used for authorization middleware
	FindByID(ctx context.Context, uid uuid.UUID) (*entity.User, error)
	// Updates user's info
	Update(ctx context.Context, user *entity.User) error
	// Deletes user
	Delete(ctx context.Context, uid uuid.UUID) error
}

type HabitsRepositoryI interface {
	// Creates new habits in database. In habit only Title, UserID, Description are necessary
	Create(ctx context.Context, habit *entity.Habit) (uuid.UUID, error)
	// Searches habit with given id
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Habit, error)
	// Lists habits owned by user with uid. Requires pagination params provided
	GetByUserID(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*entity.Habit, error)
	// Updates habit by ID (ID in habit is necessary)
	Update(ctx context.Context, habit *entity.Habit) error
	// Deletes habit with id
	Delete(ctx context.Context, id uuid.UUID) error
}

type HabitChecksRepositoryI interface {
	// Creates new check on habit with habitID
	Create(ctx context.Context, habitID uuid.UUID, date time.Time) error
	// Deletes check on habit with habitID (uncheck)
	Delete(ctx context.Context, habitID uuid.UUID, date time.Time) error
	// Inspects if check exists
	Exists(ctx context.Context, habitID uuid.UUID, date time.Time) (bool, error)
	// Provides checks of habitID for a period
	GetByHabitAndDateRange(ctx context.Context, habitID uuid.UUID, from, to time.Time) ([]entity.HabitCheck, error)
	// Returns date of last check on habitID
	GetLastCheckDate(ctx context.Context, habitID uuid.UUID) (*time.Time, error)
	// Returns count of checks for habitID
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
