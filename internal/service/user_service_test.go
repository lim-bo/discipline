package service_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/internal/service"
	"github.com/limbo/discipline/pkg/entity"
	"github.com/pressly/goose"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
)

func TestUserServiceIntegrational(t *testing.T) {
	dbCfg := setupUsersTestDB(t)
	repo := repository.NewUsersRepo(dbCfg)
	us := service.NewUserService(repo)
	ctx := context.Background()
	username := "test_user"
	password := "test_password"
	var user *entity.User
	var err error
	t.Run("registered user", func(t *testing.T) {
		user, err = us.Register(ctx, &service.RegisterRequest{
			Name:     username,
			Password: password,
		})
		assert.NoError(t, err)
		assert.Equal(t, username, user.Name)
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)))
	})
	t.Run("error registering already existed user", func(t *testing.T) {
		_, err = us.Register(ctx, &service.RegisterRequest{
			Name:     username,
			Password: password,
		})
		assert.Error(t, err)
	})
	t.Run("login", func(t *testing.T) {
		res, err := us.Login(ctx, username, password)
		assert.NoError(t, err)
		assert.Equal(t, *user, *res)
	})
	t.Run("error login on unexisted user", func(t *testing.T) {
		_, err := us.Login(ctx, "aaaaaaa", "bbbbb")
		assert.Error(t, err)
	})
	t.Run("found by name", func(t *testing.T) {
		res, err := us.GetByName(ctx, username)
		assert.NoError(t, err)
		assert.Equal(t, *user, *res)
	})
	t.Run("not found by name", func(t *testing.T) {
		_, err := us.GetByName(ctx, "unexisted")
		assert.Error(t, err)
	})
	t.Run("found by id", func(t *testing.T) {
		res, err := us.GetByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, *user, *res)
	})
	t.Run("not found by id", func(t *testing.T) {
		_, err := us.GetByID(ctx, uuid.New())
		assert.Error(t, err)
	})
	t.Run("failed to delete w/ wrong password", func(t *testing.T) {
		err := us.DeleteAccount(ctx, user.ID, "dasdasd")
		assert.Error(t, err)
	})
	t.Run("deleted", func(t *testing.T) {
		err := us.DeleteAccount(ctx, user.ID, password)
		assert.NoError(t, err)
	})
	t.Run("failed to delete unexist user", func(t *testing.T) {
		err := us.DeleteAccount(ctx, user.ID, password)
		assert.Error(t, err)
	})
}

func TestMain(m *testing.M) {
	service.InitValidator()
	m.Run()
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
