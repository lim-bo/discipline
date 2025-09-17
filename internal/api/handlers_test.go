package api_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/limbo/discipline/internal/api"
	"github.com/limbo/discipline/internal/service"
	"github.com/limbo/discipline/pkg/entity"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

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
