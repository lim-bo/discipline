package api_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/limbo/discipline/internal/api"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/mocks"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/internal/service"
	"github.com/limbo/discipline/pkg/entity"
	jwtservice "github.com/limbo/discipline/pkg/jwt_service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateHabit(t *testing.T) {
	ctrl := gomock.NewController(t)
	hService := mocks.NewMockHabitsService(ctrl)
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
	hService := mocks.NewMockHabitsService(ctrl)
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
	hService := mocks.NewMockHabitsService(ctrl)
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
	cfg := setupUsersTestDB(t)
	usersRepo := repository.NewUsersRepo(cfg)
	habitsRepo := repository.NewHabitsRepo(cfg)
	usersService := service.NewUserService(usersRepo)
	habitsService := service.NewHabitsService(habitsRepo)
	server := api.New(&api.ServicesList{
		UserService:   usersService,
		HabitsService: habitsService,
		JwtService:    jwtservice.New("secret"),
	})
	body, err := sonic.ConfigDefault.Marshal(api.RegisterRequest{
		Name:     username,
		Password: password,
	})
	if err != nil {
		t.Fatal(err)
	}
	var uid uuid.UUID
	serverAddr := "localhost:9090"
	address := "http://" + serverAddr
	go func() {
		err = server.Run(serverAddr)
		require.NoError(t, err)
	}()
	time.Sleep(time.Millisecond * 100)
	t.Run("registering new user", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, address+"/api/v1/auth/register", bytes.NewReader(body))
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		result := make(map[string]any)
		err = sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		uidStr, ok := result["uid"].(string)
		if ok {
			uid = uuid.MustParse(uidStr)
		} else {
			t.Error("invalid response body")
		}
	})
	var token string
	t.Run("logging in", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, address+"/api/v1/auth/login", bytes.NewReader(body))
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Getting auth token from response body
		result := make(map[string]any)
		defer resp.Body.Close()
		err = sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		token = "Bearer " + result["token"].(string)
	})
	habitReq := api.CreateHabitRequest{
		Title:       "test_habit",
		Description: "test_desc",
	}
	body, err = sonic.ConfigDefault.Marshal(habitReq)
	require.NoError(t, err)
	var habitID uuid.UUID
	t.Run("creating habit", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, address+"/api/v1/habits", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		defer resp.Body.Close()
		result := make(map[string]any)
		err = sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&result)
		habitID = uuid.MustParse(result["habit_id"].(string))
	})
	t.Run("getting habit", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, address+"/api/v1/habits", bytes.NewReader(body))
		require.NoError(t, err)

		req.Header.Set("Authorization", token)
		q := req.URL.Query()
		q.Add("limit", "1")
		q.Add("page", "1")
		req.URL.RawQuery = q.Encode()

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()
		var result api.GetHabitsResponse
		err = sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, 1, len(result.Habits))
		assert.Equal(t, uid.String(), result.UserID)
	})
	t.Run("deleting habit", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, address+"/api/v1/habits/"+habitID.String(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
