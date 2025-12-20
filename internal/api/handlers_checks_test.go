package api_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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

var (
	habitID = uuid.New()
)

func TestCheckHabit(t *testing.T) {
	ctrl := gomock.NewController(t)

	checksSvc := mocks.NewMockHabitChecksService(ctrl)

	serv := api.New(&api.ServicesList{
		HabitChecksService: checksSvc,
	})

	r := httptest.NewRequest(http.MethodPost, "/api/v1/habits/check", nil)
	ctx := context.WithValue(r.Context(), "User-ID", userID)
	r = r.WithContext(ctx)
	r.SetPathValue("id", habitID.String())

	date, err := time.Parse(time.DateOnly, "2025-12-18")
	require.NoError(t, err)

	q := r.URL.Query()
	q.Set("date", date.Format(time.DateOnly))

	r.URL.RawQuery = q.Encode()

	testCases := []struct {
		caseName     string
		statusCode   int
		mockPrepFunc func()
	}{
		{
			caseName:   "success",
			statusCode: http.StatusCreated,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					CheckHabit(gomock.Any(), habitID, userID, gomock.Any()).
					Return(nil)
			},
		},
		{
			caseName:   "error: habit not found",
			statusCode: http.StatusNotFound,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					CheckHabit(gomock.Any(), habitID, userID, gomock.Any()).
					Return(errorvalues.ErrHabitNotFound)
			},
		},
		{
			caseName:   "error: wrong owner",
			statusCode: http.StatusNotFound,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					CheckHabit(gomock.Any(), habitID, userID, gomock.Any()).
					Return(errorvalues.ErrWrongOwner)
			},
		},
		{
			caseName:   "error: check exists",
			statusCode: http.StatusConflict,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					CheckHabit(gomock.Any(), habitID, userID, gomock.Any()).
					Return(errorvalues.ErrCheckExist)
			},
		},
		{
			caseName:   "error: internal",
			statusCode: http.StatusInternalServerError,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					CheckHabit(gomock.Any(), habitID, userID, gomock.Any()).
					Return(errors.New("service error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			w := httptest.NewRecorder()
			tc.mockPrepFunc()
			serv.CheckHabit(w, r)
			assert.Equal(t, tc.statusCode, w.Result().StatusCode)
		})
	}
}

func TestUncheckHabit(t *testing.T) {
	ctrl := gomock.NewController(t)

	checksSvc := mocks.NewMockHabitChecksService(ctrl)

	serv := api.New(&api.ServicesList{
		HabitChecksService: checksSvc,
	})

	r := httptest.NewRequest(http.MethodDelete, "/api/v1/habits/check", nil)
	ctx := context.WithValue(r.Context(), "User-ID", userID)
	r = r.WithContext(ctx)
	r.SetPathValue("id", habitID.String())

	date, err := time.Parse(time.DateOnly, "2025-12-18")
	require.NoError(t, err)

	q := r.URL.Query()
	q.Set("date", date.Format(time.DateOnly))

	r.URL.RawQuery = q.Encode()

	testCases := []struct {
		caseName     string
		statusCode   int
		mockPrepFunc func()
	}{
		{
			caseName:   "success",
			statusCode: http.StatusNoContent,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					UncheckHabit(gomock.Any(), habitID, userID, gomock.Any()).
					Return(nil)
			},
		},
		{
			caseName:   "error: habit not found",
			statusCode: http.StatusNotFound,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					UncheckHabit(gomock.Any(), habitID, userID, gomock.Any()).
					Return(errorvalues.ErrHabitNotFound)
			},
		},
		{
			caseName:   "error: wrong owner",
			statusCode: http.StatusNotFound,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					UncheckHabit(gomock.Any(), habitID, userID, gomock.Any()).
					Return(errorvalues.ErrWrongOwner)
			},
		},
		{
			caseName:   "error: check doesn't exists",
			statusCode: http.StatusConflict,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					UncheckHabit(gomock.Any(), habitID, userID, gomock.Any()).
					Return(errorvalues.ErrCheckNotFound)
			},
		},
		{
			caseName:   "error: internal",
			statusCode: http.StatusInternalServerError,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					UncheckHabit(gomock.Any(), habitID, userID, gomock.Any()).
					Return(errors.New("service error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			w := httptest.NewRecorder()
			tc.mockPrepFunc()
			serv.UncheckHabit(w, r)
			assert.Equal(t, tc.statusCode, w.Result().StatusCode)
		})
	}
}

func TestHabitStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	checksSvc := mocks.NewMockHabitChecksService(ctrl)

	serv := api.New(&api.ServicesList{
		HabitChecksService: checksSvc,
	})

	r := httptest.NewRequest(http.MethodGet, "/api/v1/habits/stat/"+habitID.String(), nil)
	r.SetPathValue("id", habitID.String())
	ctx := context.WithValue(r.Context(), "User-ID", userID)
	r = r.WithContext(ctx)

	date := time.Now()
	stats := entity.HabitStats{
		ID:            habitID,
		TotalChecks:   15,
		CurrentStreak: 10,
		MaxStreak:     10,
		LastCheck:     date,
	}

	testCases := []struct {
		caseName     string
		statusCode   int
		result       entity.HabitStats
		mockPrepFunc func()
	}{
		{
			caseName:   "success",
			statusCode: http.StatusOK,
			result:     stats,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					GetHabitStats(gomock.Any(), habitID, userID).
					Return(&stats, nil)
			},
		},
		{
			caseName:   "error: not found",
			statusCode: http.StatusNotFound,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					GetHabitStats(gomock.Any(), habitID, userID).
					Return(nil, errorvalues.ErrHabitNotFound)
			},
		},
		{
			caseName:   "error: wrong owner",
			statusCode: http.StatusNotFound,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					GetHabitStats(gomock.Any(), habitID, userID).
					Return(nil, errorvalues.ErrWrongOwner)
			},
		},
		{
			caseName:   "error: internal",
			statusCode: http.StatusInternalServerError,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					GetHabitStats(gomock.Any(), habitID, userID).
					Return(nil, errors.New("service error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.mockPrepFunc()
			w := httptest.NewRecorder()
			serv.GetHabitStats(w, r)
			assert.Equal(t, tc.statusCode, w.Result().StatusCode)
			if tc.statusCode == http.StatusOK && w.Result().StatusCode == http.StatusOK {
				var statsResponse entity.HabitStats
				err := sonic.ConfigDefault.NewDecoder(w.Result().Body).Decode(&statsResponse)
				require.NoError(t, err)
				assert.Equal(t, tc.result.ID, statsResponse.ID)
				assert.Equal(t, tc.result.CurrentStreak, statsResponse.CurrentStreak)
				assert.Equal(t, tc.result.MaxStreak, statsResponse.MaxStreak)
				assert.Equal(t, tc.result.TotalChecks, statsResponse.TotalChecks)
				assert.True(t, tc.result.LastCheck.Equal(statsResponse.LastCheck))
			}
		})
	}
}

func TestGetChecks(t *testing.T) {
	ctrl := gomock.NewController(t)
	checksSvc := mocks.NewMockHabitChecksService(ctrl)

	serv := api.New(&api.ServicesList{
		HabitChecksService: checksSvc,
	})

	from, err := time.Parse(time.DateOnly, "2025-12-01")
	require.NoError(t, err)
	now := time.Now()
	to := now.Truncate(time.Hour * 24).UTC()

	r := httptest.NewRequest(http.MethodGet, "/api/v1/habits/check/"+habitID.String(), nil)
	r.SetPathValue("id", habitID.String())
	ctx := context.WithValue(r.Context(), "User-ID", userID)
	r = r.WithContext(ctx)
	q := r.URL.Query()
	q.Add("from", from.Format(time.DateOnly))
	q.Add("to", to.Format(time.DateOnly))
	r.URL.RawQuery = q.Encode()

	checks := make([]entity.HabitCheck, 0, 5)
	for i := range 5 {
		checks = append(checks, entity.HabitCheck{
			ID:        i,
			HabitID:   habitID,
			CheckDate: now.Add(time.Hour * 24 * time.Duration(-i)),
			CreatedAt: now.Add(time.Hour * 24 * time.Duration(-i)),
		})
	}

	testCases := []struct {
		caseName     string
		statusCode   int
		result       []entity.HabitCheck
		mockPrepFunc func()
	}{
		{
			caseName:   "success",
			statusCode: http.StatusOK,
			result:     checks,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					GetHabitChecks(gomock.Any(), habitID, userID, from, to).
					Return(checks, nil)
			},
		},
		{
			caseName:   "error: habit not found",
			statusCode: http.StatusNotFound,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					GetHabitChecks(gomock.Any(), habitID, userID, from, to).
					Return(nil, errorvalues.ErrHabitNotFound)
			},
		},
		{
			caseName:   "error: wrong owner",
			statusCode: http.StatusNotFound,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					GetHabitChecks(gomock.Any(), habitID, userID, from, to).
					Return(nil, errorvalues.ErrWrongOwner)
			},
		},
		{
			caseName:   "error: internal",
			statusCode: http.StatusInternalServerError,
			mockPrepFunc: func() {
				checksSvc.EXPECT().
					GetHabitChecks(gomock.Any(), habitID, userID, from, to).
					Return(nil, errors.New("service error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.mockPrepFunc()
			w := httptest.NewRecorder()
			serv.GetChecks(w, r)
			assert.Equal(t, tc.statusCode, w.Result().StatusCode)
			if tc.statusCode == http.StatusOK && w.Result().StatusCode == http.StatusOK {
				var resp api.GetChecksResponse
				err = sonic.ConfigDefault.NewDecoder(w.Result().Body).Decode(&resp)
				for i := range resp.Values {
					assert.Equal(t, checks[i].ID, resp.Values[i].ID)
					assert.Equal(t, checks[i].HabitID, resp.Values[i].HabitID)
					assert.True(t, checks[i].CheckDate.Equal(resp.Values[i].CheckDate))
					assert.True(t, checks[i].CreatedAt.Equal(resp.Values[i].CreatedAt))
				}
			}
		})
	}
}

func TestChecksIntegral(t *testing.T) {
	cfg := setupTestDB(t)
	habitRepo := repository.NewHabitsRepo(cfg)
	checksRepo := repository.NewHabitChecksRepo(cfg)
	usersRepo := repository.NewUsersRepo(cfg)
	checksSvc := service.NewHabitChecksService(habitRepo, checksRepo)
	habitsSvc := service.NewHabitsService(habitRepo)
	usersSvc := service.NewUserService(usersRepo)

	server := api.New(&api.ServicesList{
		HabitsService:      habitsSvc,
		HabitChecksService: checksSvc,
		UserService:        usersSvc,
		JwtService:         jwtservice.New("secret"),
	})

	body, err := sonic.ConfigDefault.Marshal(api.RegisterRequest{
		Name:     username,
		Password: password,
	})
	if err != nil {
		t.Fatal(err)
	}
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

	t.Run("checking habit", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, address+"/api/v1/habits/check/"+habitID.String(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("unchecking habit", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, address+"/api/v1/habits/check/"+habitID.String(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("checking habit to few dates", func(t *testing.T) {
		now := time.Now().Truncate(time.Hour * 24)
		for i := range 5 {
			date := now.Add(time.Hour * 24 * -time.Duration(i))
			req, err := http.NewRequest(http.MethodPost, address+"/api/v1/habits/check/"+habitID.String(), nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", token)
			q := req.URL.Query()
			q.Set("date", date.Format(time.DateOnly))
			req.URL.RawQuery = q.Encode()

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		}
	})

	t.Run("getting checks", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, address+"/api/v1/habits/check/"+habitID.String(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()
		var checks api.GetChecksResponse
		err = sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&checks)
		require.NoError(t, err)
		assert.Equal(t, 5, len(checks.Values))
	})

	t.Run("getting stats", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, address+"/api/v1/habits/stat/"+habitID.String(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()
		var stats entity.HabitStats
		err = sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&stats)
		require.NoError(t, err)
		assert.Equal(t, 5, stats.TotalChecks)
		assert.Equal(t, 5, stats.CurrentStreak)
		assert.Equal(t, 5, stats.MaxStreak)
		assert.Equal(t, habitID, stats.ID)
	})
}
