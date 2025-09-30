package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/limbo/discipline/internal/api"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/internal/service"
	"github.com/limbo/discipline/internal/service/mocks"
	"github.com/limbo/discipline/pkg/entity"
	jwtservice "github.com/limbo/discipline/pkg/jwt_service"
	"github.com/pressly/goose"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func testHandler(w http.ResponseWriter, r *http.Request) {
	uid, err := api.GetUIDFromContext(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"uid": ` + uid.String() + `}`))
}

func TestAuthMiddleware(t *testing.T) {
	secret := "secret"
	cfg := setupUsersTestDB(t)
	repo := repository.NewUsersRepo(cfg)
	userService := service.NewUserService(repo)
	serv := api.New(&api.ServicesList{
		UserService: userService,
		JwtService:  jwtservice.New(secret),
	})
	handler := serv.AuthMiddleware(http.HandlerFunc(testHandler))
	// Creating user to login
	body, err := sonic.ConfigDefault.Marshal(api.RegisterRequest{
		Name:     username,
		Password: password,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Run("creating user", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		serv.Register(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Result().StatusCode)
	})
	var token string
	var ok bool
	t.Run("logging in and getting token", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		serv.Login(rr, req)
		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
		result := make(map[string]any)
		err := sonic.ConfigDefault.NewDecoder(rr.Result().Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}
		token, ok = result["token"].(string)
		if !ok || token == "" {
			t.Error("invalid token")
		}
		t.Log("token: ", token)
	})
	t.Run("successful auth", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/endpoint", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	})
	t.Run("error", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/endpoint", nil)
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
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

var (
	userID = uuid.New()
)

func TestCreateHabit(t *testing.T) {
	ctrl := gomock.NewController(t)
	hService := mocks.NewMockHabitsServiceI(ctrl)
	serv := api.New(&api.ServicesList{
		HabitsService: hService,
	})
	habit := api.CreateHabitRequest{
		Title:       "test_habit",
		Description: "test_habit_description",
	}
	body, err := sonic.ConfigDefault.Marshal(habit)
	require.NoError(t, err)
	habitID := uuid.New()

	testCases := []struct {
		ExpectedCode int
		MockPrepFunc func()
		Body         io.Reader
	}{
		{
			ExpectedCode: http.StatusCreated,
			MockPrepFunc: func() {
				hService.EXPECT().CreateHabit(gomock.Any(), userID, service.CreateHabitRequest{
					Title:       habit.Title,
					Description: habit.Description,
				}).Return(&entity.Habit{
					ID:          habitID,
					UserID:      uid,
					Title:       habit.Title,
					Description: habit.Description,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, nil)
			},
			Body: bytes.NewReader(body),
		},
		{
			ExpectedCode: http.StatusConflict,
			MockPrepFunc: func() {
				hService.EXPECT().CreateHabit(gomock.Any(), userID, service.CreateHabitRequest{
					Title:       habit.Title,
					Description: habit.Description,
				}).Return(nil, errorvalues.ErrUserHasHabit)
			},
			Body: bytes.NewReader(body),
		},
		{
			ExpectedCode: http.StatusNotFound,
			MockPrepFunc: func() {
				hService.EXPECT().CreateHabit(gomock.Any(), userID, service.CreateHabitRequest{
					Title:       habit.Title,
					Description: habit.Description,
				}).Return(nil, errorvalues.ErrUserNotFound)
			},
			Body: bytes.NewReader(body),
		},
		{
			ExpectedCode: http.StatusInternalServerError,
			MockPrepFunc: func() {
				hService.EXPECT().CreateHabit(gomock.Any(), userID, service.CreateHabitRequest{
					Title:       habit.Title,
					Description: habit.Description,
				}).Return(nil, errors.New("service error"))
			},
			Body: bytes.NewReader(body),
		},
		{
			ExpectedCode: http.StatusBadRequest,
			MockPrepFunc: func() {},
			Body:         bytes.NewReader([]byte("corrupted")),
		},
	}
	for _, tc := range testCases {
		tc.MockPrepFunc()
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/habits", tc.Body)
		r = r.WithContext(context.WithValue(r.Context(), "User-ID", userID))
		serv.CreateHabit(rr, r)
		assert.Equal(t, tc.ExpectedCode, rr.Result().StatusCode)
		if tc.ExpectedCode == http.StatusCreated {
			resp, _ := io.ReadAll(rr.Result().Body)
			fmt.Println(string(resp))
		}
	}
}
func TestGetHabits(t *testing.T) {
	ctrl := gomock.NewController(t)
	hService := mocks.NewMockHabitsServiceI(ctrl)
	serv := api.New(&api.ServicesList{
		HabitsService: hService,
	})
	habits := make([]*entity.Habit, 0, 10)
	for i := range 10 {
		habits = append(habits, &entity.Habit{
			ID:          uuid.New(),
			UserID:      userID,
			Title:       fmt.Sprintf("test_habit_%d", i+1),
			Description: "blah blah blah",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}
	testCases := []struct {
		ExpectedCode        int
		MockPrepFunc        func()
		Limit               int
		Page                int
		ExpectedHabitsCount int
	}{
		{
			ExpectedCode: http.StatusOK,
			MockPrepFunc: func() {
				hService.EXPECT().GetUserHabits(gomock.Any(), userID, service.PaginationOpts{
					Limit:  10,
					Offset: 0,
				}).Return(habits, nil)
			},
			Page:                1,
			Limit:               10,
			ExpectedHabitsCount: 10,
		},
		{
			ExpectedCode: http.StatusOK,
			MockPrepFunc: func() {
				hService.EXPECT().GetUserHabits(gomock.Any(), userID, service.PaginationOpts{
					Limit:  4,
					Offset: 4,
				}).Return(habits[2:6], nil)
			},
			Page:                2,
			Limit:               4,
			ExpectedHabitsCount: 4,
		},
		{
			ExpectedCode: http.StatusInternalServerError,
			MockPrepFunc: func() {
				hService.EXPECT().GetUserHabits(gomock.Any(), userID, service.PaginationOpts{
					Limit:  10,
					Offset: 0,
				}).Return(nil, errors.New("service error"))
			},
			Page:                1,
			Limit:               10,
			ExpectedHabitsCount: 0,
		},
	}
	for _, tc := range testCases {
		tc.MockPrepFunc()
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/habits", nil)
		q := r.URL.Query()
		q.Add("limit", strconv.Itoa(tc.Limit))
		q.Add("page", strconv.Itoa(tc.Page))
		r.URL.RawQuery = q.Encode()
		r = r.WithContext(context.WithValue(r.Context(), "User-ID", userID))
		serv.GetHabits(rr, r)
		assert.Equal(t, tc.ExpectedCode, rr.Result().StatusCode)
		if rr.Result().StatusCode == http.StatusOK {
			var resp api.GetHabitsResponse
			err := sonic.ConfigDefault.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)
			assert.Equal(t, tc.ExpectedHabitsCount, len(resp.Habits))
		}
	}
}
func TestDeleteHabit(t *testing.T) {
	ctrl := gomock.NewController(t)
	hService := mocks.NewMockHabitsServiceI(ctrl)
	serv := api.New(&api.ServicesList{
		HabitsService: hService,
	})
	habitID := uuid.New()
	testCases := []struct {
		ExpectedCode int
		MockPrepFunc func()
	}{
		{
			ExpectedCode: http.StatusOK,
			MockPrepFunc: func() {
				hService.EXPECT().DeleteHabit(gomock.Any(), habitID, userID).Return(nil)
			},
		},
		{
			ExpectedCode: http.StatusNotFound,
			MockPrepFunc: func() {
				hService.EXPECT().DeleteHabit(gomock.Any(), habitID, userID).Return(errorvalues.ErrHabitNotFound)
			},
		},
		{
			ExpectedCode: http.StatusNotFound,
			MockPrepFunc: func() {
				hService.EXPECT().DeleteHabit(gomock.Any(), habitID, userID).Return(errorvalues.ErrWrongOwner)
			},
		},
		{
			ExpectedCode: http.StatusInternalServerError,
			MockPrepFunc: func() {
				hService.EXPECT().DeleteHabit(gomock.Any(), habitID, userID).Return(errors.New("service error"))
			},
		},
	}
	for _, tc := range testCases {
		tc.MockPrepFunc()
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/api/habits/"+habitID.String(), nil)
		r = r.WithContext(context.WithValue(r.Context(), "User-ID", userID))
		r.SetPathValue("id", habitID.String())
		serv.DeleteHabit(rr, r)
		assert.Equal(t, tc.ExpectedCode, rr.Result().StatusCode)
	}
}
func TestHabitsCRUDIntegrational(t *testing.T) {

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
