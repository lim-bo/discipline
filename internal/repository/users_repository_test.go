package repository_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/pkg/entity"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	conn, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	user := entity.User{
		Name:         "test_user",
		PasswordHash: "test_password_hash",
	}
	query := regexp.QuoteMeta(`INSERT INTO users (name, password_hash) VALUES ($1, $2);`)
	ctx := context.Background()
	repo := repository.NewUsersRepoWithConn(conn)
	t.Run("successfully created", func(t *testing.T) {
		conn.ExpectExec(query).WithArgs(user.Name, user.PasswordHash).WillReturnResult(pgxmock.NewResult("INSERT", 1))
		err := repo.Create(ctx, &user)
		assert.NoError(t, err)
	})
	t.Run("unique violation error", func(t *testing.T) {
		conn.ExpectExec(query).WithArgs(user.Name, user.PasswordHash).WillReturnError(&pgconn.PgError{
			Code: "23505",
		})
		err := repo.Create(ctx, &user)
		assert.ErrorIs(t, err, errorvalues.ErrUserExists)
	})
	t.Run("db error", func(t *testing.T) {
		conn.ExpectExec(query).WithArgs(user.Name, user.PasswordHash).WillReturnError(errors.New("db error"))
		err := repo.Create(ctx, &user)
		assert.Error(t, err)
	})
}

func TestFindByName(t *testing.T) {
	conn, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	repo := repository.NewUsersRepoWithConn(conn)
	user := entity.User{
		ID:           uuid.New(),
		Name:         "test_user",
		PasswordHash: "test_password_hash",
	}
	query := regexp.QuoteMeta(`SELECT id, name, password_hash FROM users WHERE name = $1;`)
	t.Run("found", func(t *testing.T) {
		conn.ExpectQuery(query).
			WithArgs(user.Name).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "password_hash"}).AddRow(user.ID, user.Name, user.PasswordHash))
		result, err := repo.FindByName(ctx, user.Name)
		assert.NoError(t, err)
		assert.Equal(t, user, *result)
	})
	t.Run("not found", func(t *testing.T) {
		conn.ExpectQuery(query).
			WithArgs(user.Name).
			WillReturnError(pgx.ErrNoRows)
		_, err := repo.FindByName(ctx, user.Name)
		assert.ErrorIs(t, err, errorvalues.ErrUserNotFound)
	})
	t.Run("db error", func(t *testing.T) {
		conn.ExpectQuery(query).
			WithArgs(user.Name).
			WillReturnError(errors.New("db error"))
		_, err := repo.FindByName(ctx, user.Name)
		assert.Error(t, err)
	})
}

func TestFindByID(t *testing.T) {
	conn, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	repo := repository.NewUsersRepoWithConn(conn)
	user := entity.User{
		ID:           uuid.New(),
		Name:         "test_user",
		PasswordHash: "test_password_hash",
	}
	query := regexp.QuoteMeta(`SELECT id, name, password_hash FROM users WHERE id = $1;`)
	t.Run("found", func(t *testing.T) {
		conn.ExpectQuery(query).
			WithArgs(user.ID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "password_hash"}).AddRow(user.ID, user.Name, user.PasswordHash))
		result, err := repo.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, user, *result)
	})
	t.Run("not found", func(t *testing.T) {
		conn.ExpectQuery(query).
			WithArgs(user.ID).
			WillReturnError(pgx.ErrNoRows)
		_, err := repo.FindByID(ctx, user.ID)
		assert.ErrorIs(t, err, errorvalues.ErrUserNotFound)
	})
	t.Run("db error", func(t *testing.T) {
		conn.ExpectQuery(query).
			WithArgs(user.ID).
			WillReturnError(errors.New("db error"))
		_, err := repo.FindByID(ctx, user.ID)
		assert.Error(t, err)
	})
}

func TestUpdateUser(t *testing.T) {
	conn, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	repo := repository.NewUsersRepoWithConn(conn)
	user := entity.User{
		ID:           uuid.New(),
		Name:         "test_user",
		PasswordHash: "test_password_hash",
	}
	query := regexp.QuoteMeta(`UPDATE users SET name = $1, password_hash = $2 WHERE id = $3;`)
	t.Run("updated", func(t *testing.T) {
		conn.ExpectExec(query).
			WithArgs(user.Name, user.PasswordHash, user.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		err := repo.Update(ctx, &user)
		assert.NoError(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		conn.ExpectExec(query).
			WithArgs(user.Name, user.PasswordHash, user.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))
		err := repo.Update(ctx, &user)
		assert.ErrorIs(t, err, errorvalues.ErrUserNotFound)
	})
	t.Run("db error", func(t *testing.T) {
		conn.ExpectExec(query).
			WithArgs(user.Name, user.PasswordHash, user.ID).
			WillReturnError(errors.New("db error"))
		err := repo.Update(ctx, &user)
		assert.Error(t, err)
	})
}

func TestDeleteUser(t *testing.T) {
	conn, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	repo := repository.NewUsersRepoWithConn(conn)
	uid := uuid.New()
	query := regexp.QuoteMeta(`DELETE FROM users WHERE id = $1;`)
	t.Run("deleted", func(t *testing.T) {
		conn.ExpectExec(query).
			WithArgs(uid).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		err := repo.Delete(ctx, uid)
		assert.NoError(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		conn.ExpectExec(query).
			WithArgs(uid).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))
		err := repo.Delete(ctx, uid)
		assert.ErrorIs(t, err, errorvalues.ErrUserNotFound)
	})
	t.Run("db error", func(t *testing.T) {
		conn.ExpectExec(query).
			WithArgs(uid).
			WillReturnError(errors.New("db error"))
		err := repo.Delete(ctx, uid)
		assert.Error(t, err)
	})
}
