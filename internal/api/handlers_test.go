package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/limbo/discipline/internal/api"
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

func TestMain(m *testing.M) {
	service.InitValidator()
	m.Run()
}

type UserServiceMock struct {
	success bool
}

func (usmock *UserServiceMock) ChangeState(success bool) {
	usmock.success = success
}

func (usmock *UserServiceMock) Register(ctx context.Context, req *service.RegisterRequest) (*entity.User, error) {
	if usmock.success {
		return &entity.User{
			ID:           uid,
			Name:         username,
			PasswordHash: string(passwordHash),
		}, nil
	}
	return nil, errors.New("mocked error")
}

func (usmock *UserServiceMock) Login(ctx context.Context, name, password string) (*entity.User, error) {
	if usmock.success {
		return &entity.User{
			ID:           uid,
			Name:         username,
			PasswordHash: string(passwordHash),
		}, nil
	}
	return nil, errors.New("mocked error")
}

func (usmock *UserServiceMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	if usmock.success {
		return &entity.User{
			ID:           uid,
			Name:         username,
			PasswordHash: string(passwordHash),
		}, nil
	}
	return nil, errors.New("mocked error")
}
func (usmock *UserServiceMock) GetByName(ctx context.Context, name string) (*entity.User, error) {
	if usmock.success {
		return &entity.User{
			ID:           uid,
			Name:         username,
			PasswordHash: string(passwordHash),
		}, nil
	}
	return nil, errors.New("mocked error")
}
func (usmock *UserServiceMock) DeleteAccount(ctx context.Context, id uuid.UUID, password string) error {
	if usmock.success {
		return nil
	}
	return errors.New("mocked error")
}

var (
	username        = "test_name"
	password        = "test_password"
	passwordHash, _ = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	uid             = uuid.New()
)

func TestRegister(t *testing.T) {
	body, err := sonic.ConfigDefault.Marshal(api.RegisterRequest{
		Name:     username,
		Password: password,
	})
	if err != nil {
		t.Fatal(err)
	}
	var req *http.Request
	mock := UserServiceMock{}
	serv := api.New(&api.ServicesList{
		UserService: &mock,
	})
	t.Run("registered", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		mock.ChangeState(true)
		serv.Register(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Result().StatusCode)
	})

	t.Run("service error", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		mock.ChangeState(false)
		serv.Register(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Result().StatusCode)
	})

	t.Run("invalid body", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/auth/register", nil)
		mock.ChangeState(true)
		serv.Login(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
	})
}

func TestLogin(t *testing.T) {
	body, err := sonic.ConfigDefault.Marshal(api.LoginRequest{
		Name:     username,
		Password: password,
	})
	if err != nil {
		t.Fatal(err)
	}
	var req *http.Request
	mock := UserServiceMock{}
	serv := api.New(&api.ServicesList{
		UserService: &mock,
	})
	t.Run("logged in", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		mock.ChangeState(true)
		serv.Login(rr, req)
		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	})
	t.Run("invalid body", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		mock.ChangeState(true)
		serv.Login(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
	})
	t.Run("service error", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		mock.ChangeState(false)
		serv.Login(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Result().StatusCode)
	})
}

func TestUsersHandlersIntegrational(t *testing.T) {
	cfg := setupUsersTestDB(t)
	repo := repository.NewUsersRepo(cfg)
	userService := service.NewUserService(repo)
	server := api.New(&api.ServicesList{
		UserService: userService,
	})
	body, err := sonic.ConfigDefault.Marshal(api.RegisterRequest{
		Name:     username,
		Password: password,
	})
	if err != nil {
		t.Fatal(err)
	}
	var uid uuid.UUID
	t.Run("successfully registered", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		server.Register(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Result().StatusCode)
		result := make(map[string]any)
		err := sonic.ConfigDefault.NewDecoder(rr.Result().Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}
		defer rr.Result().Body.Close()
		uidStr, ok := result["uid"].(string)
		if ok {
			uid = uuid.MustParse(uidStr)
		} else {
			t.Error("invalid response body")
		}
	})
	t.Run("error registered: invalid body", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/register", nil)
		server.Register(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
	})
	t.Run("successfully logged in", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		server.Login(rr, req)
		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
		result := make(map[string]any)
		err := sonic.ConfigDefault.NewDecoder(rr.Result().Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}
		defer rr.Result().Body.Close()
		uidStr, ok := result["uid"].(string)
		var uidLogin uuid.UUID
		if ok {
			uidLogin = uuid.MustParse(uidStr)
		} else {
			t.Error("invalid response body")
		}
		assert.Equal(t, uid, uidLogin)
	})
	t.Run("error login: invalid body", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		server.Register(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
	})
	t.Run("error login: wrong password", func(t *testing.T) {
		body, err := sonic.ConfigDefault.Marshal(api.RegisterRequest{
			Name:     username,
			Password: password + "12345",
		})
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		server.Login(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Result().StatusCode)
	})
	t.Run("error login: user not found", func(t *testing.T) {
		body, err := sonic.ConfigDefault.Marshal(api.RegisterRequest{
			Name:     username + "dasdwdasd",
			Password: password,
		})
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		server.Login(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Result().StatusCode)
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
