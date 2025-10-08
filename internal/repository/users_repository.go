package repository

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/pkg/cleanup"
	"github.com/limbo/discipline/pkg/entity"
)

type UsersRepository struct {
	conn PgConnection
}

func NewUsersRepo(cfg DBConfig) *UsersRepository {
	pool, err := pgxpool.New(context.Background(), cfg.ConnString())
	if err != nil {
		log.Fatal("creating connection for usersRepo error: " + err.Error())
	}
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal("error while pinging connection for usersRepo: " + err.Error())
	}
	cleanup.Register(&cleanup.Job{
		Name: "closing pgxpool",
		F: func() error {
			pool.Close()
			return nil
		},
	})
	return &UsersRepository{
		conn: pool,
	}
}

func NewUsersRepoWithConn(conn PgConnection) *UsersRepository {
	err := conn.Ping(context.Background())
	if err != nil {
		log.Fatal("error while pinging connection for usersRepo: " + err.Error())
	}
	return &UsersRepository{
		conn: conn,
	}
}

func (ur *UsersRepository) Create(ctx context.Context, user *entity.User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	_, err := ur.conn.Exec(ctx, `INSERT INTO users (name, password_hash) VALUES ($1, $2);`, user.Name, user.PasswordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			// Unique violation
			case "23505":
				return errorvalues.ErrUserExists
			}
		}
		return errors.New("creating user db error: " + err.Error())
	}
	return nil
}

func (ur *UsersRepository) FindByName(ctx context.Context, name string) (*entity.User, error) {
	var user entity.User
	row := ur.conn.QueryRow(ctx, `SELECT id, name, password_hash FROM users WHERE name = $1;`, name)
	if err := row.Scan(&user.ID, &user.Name, &user.PasswordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorvalues.ErrUserNotFound
		}
		return nil, errors.New("searching user by name error: " + err.Error())
	}
	return &user, nil
}

func (ur *UsersRepository) FindByID(ctx context.Context, uid uuid.UUID) (*entity.User, error) {
	var user entity.User
	row := ur.conn.QueryRow(ctx, `SELECT id, name, password_hash FROM users WHERE id = $1;`, uid)
	if err := row.Scan(&user.ID, &user.Name, &user.PasswordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorvalues.ErrUserNotFound
		}
		return nil, errors.New("searching user by id error: " + err.Error())
	}
	return &user, nil
}

func (ur *UsersRepository) Update(ctx context.Context, user *entity.User) error {
	ct, err := ur.conn.Exec(ctx, `UPDATE users SET name = $1, password_hash = $2 WHERE id = $3;`,
		user.Name,
		user.PasswordHash,
		user.ID,
	)
	if err != nil {
		return errors.New("updating user error: " + err.Error())
	}
	if ct.RowsAffected() == 0 {
		return errorvalues.ErrUserNotFound
	}
	return nil
}

func (ur *UsersRepository) Delete(ctx context.Context, uid uuid.UUID) error {
	ct, err := ur.conn.Exec(ctx, `DELETE FROM users WHERE id = $1;`, uid)
	if err != nil {
		return errors.New("deleting user error: " + err.Error())
	}
	if ct.RowsAffected() == 0 {
		return errorvalues.ErrUserNotFound
	}
	return nil
}
