package repository

import (
	"context"

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
