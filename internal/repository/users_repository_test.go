package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/lib/pq"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/pkg/entity"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/pressly/goose"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
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

func TestUsersIntegrational(t *testing.T) {
	cfg := setupUsersTestDB(t)
	repo := repository.NewUsersRepo(cfg)
	user := entity.User{
		Name:         "test_user",
		PasswordHash: "some_test_hash",
	}
	ctx := context.Background()
	t.Run("successfully created user", func(t *testing.T) {
		err := repo.Create(ctx, &user)
		assert.NoError(t, err)
	})
	t.Run("user found by name", func(t *testing.T) {
		res, err := repo.FindByName(ctx, user.Name)
		assert.NoError(t, err)
		user.ID = res.ID
		assert.Equal(t, user, *res)
	})
	t.Run("user not found by name", func(t *testing.T) {
		_, err := repo.FindByName(ctx, "unknown")
		assert.ErrorIs(t, err, errorvalues.ErrUserNotFound)
	})
	t.Run("user found by id", func(t *testing.T) {
		res, err := repo.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, user, *res)
	})
	t.Run("user not found by id", func(t *testing.T) {
		_, err := repo.FindByID(ctx, uuid.New())
		assert.ErrorIs(t, err, errorvalues.ErrUserNotFound)
	})
	newUserCredentials := entity.User{
		ID:           user.ID,
		Name:         "new_test_user",
		PasswordHash: "other_test_hash",
	}
	t.Run("user updated", func(t *testing.T) {
		err := repo.Update(ctx, &newUserCredentials)
		assert.NoError(t, err)
		res, err := repo.FindByID(ctx, newUserCredentials.ID)
		assert.NoError(t, err)
		assert.Equal(t, newUserCredentials, *res)
	})
	t.Run("user for update not found", func(t *testing.T) {
		err := repo.Update(ctx, &entity.User{
			ID: uuid.New(),
		})
		assert.ErrorIs(t, err, errorvalues.ErrUserNotFound)
	})
	t.Run("user for deletion not found", func(t *testing.T) {
		err := repo.Delete(ctx, uuid.New())
		assert.ErrorIs(t, err, errorvalues.ErrUserNotFound)
	})
	t.Run("user deleted", func(t *testing.T) {
		err := repo.Delete(ctx, newUserCredentials.ID)
		assert.NoError(t, err)
	})
}

type testPGConfig struct {
	connStr string
}

func (cfg *testPGConfig) ConnString() string {
	return cfg.connStr
}

func setupUsersTestDB(t *testing.T) *testPGConfig {
	container, err := postgres.Run(context.Background(), "postgres:17",
		postgres.WithUsername("test_user"),
		postgres.WithDatabase("barn"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatal("error running test container: " + err.Error())
	}
	connStr, err := container.ConnectionString(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	connStr += "sslmode=disable"
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatal(err)
	}
	err = goose.Up(conn, "../../migrations")
	if err != nil {
		t.Fatal(err)
	}

	conn.Close()
	t.Cleanup(func() {
		container.Terminate(context.Background())
	})
	return &testPGConfig{
		connStr: connStr,
	}
}
