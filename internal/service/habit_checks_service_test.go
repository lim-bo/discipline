package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/internal/repository/mocks"
	"github.com/limbo/discipline/internal/service"
	"github.com/limbo/discipline/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCheckHabit(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	checksRepo := mocks.NewMockHabitChecksRepositoryI(ctrl)
	habitsRepo := mocks.NewMockHabitsRepositoryI(ctrl)

	serv := service.NewHabitChecksService(habitsRepo, checksRepo)
	habitID := uuid.New()
	userID := uuid.New()
	checkDate := time.Now()
	testCases := []struct {
		Desc         string
		Error        error
		HabitID      uuid.UUID
		UserID       uuid.UUID
		CheckDate    time.Time
		MockPrepFunc func()
	}{
		{
			Desc:      "success",
			Error:     nil,
			HabitID:   habitID,
			UserID:    userID,
			CheckDate: checkDate,
			MockPrepFunc: func() {
				habitsRepo.EXPECT().GetByID(gomock.Any(), habitID).Return(&entity.Habit{
					ID:          habitID,
					UserID:      userID,
					Title:       "test_habit",
					Description: "test_desc",
				}, nil)
				checksRepo.EXPECT().Exists(gomock.Any(), habitID, checkDate).Return(false, nil)
				checksRepo.EXPECT().Create(gomock.Any(), habitID, checkDate).Return(nil)
			},
		},
		{
			Desc:      "error wrong owner",
			Error:     errorvalues.ErrWrongOwner,
			HabitID:   habitID,
			UserID:    userID,
			CheckDate: checkDate,
			MockPrepFunc: func() {
				habitsRepo.EXPECT().GetByID(gomock.Any(), habitID).Return(&entity.Habit{
					ID:          habitID,
					UserID:      uuid.New(),
					Title:       "test_habit",
					Description: "test_desc",
				}, nil)
			},
		},
		{
			Desc:      "error check date not allowed",
			Error:     errorvalues.ErrCheckDateNotAllowed,
			HabitID:   habitID,
			UserID:    userID,
			CheckDate: checkDate.Add(time.Hour * 72),
			MockPrepFunc: func() {
				habitsRepo.EXPECT().GetByID(gomock.Any(), habitID).Return(&entity.Habit{
					ID:          habitID,
					UserID:      userID,
					Title:       "test_habit",
					Description: "test_desc",
				}, nil)
			},
		},
		{
			Desc:      "error creating existed check",
			Error:     errorvalues.ErrCheckExist,
			HabitID:   habitID,
			UserID:    userID,
			CheckDate: checkDate,
			MockPrepFunc: func() {
				habitsRepo.EXPECT().GetByID(gomock.Any(), habitID).Return(&entity.Habit{
					ID:          habitID,
					UserID:      userID,
					Title:       "test_habit",
					Description: "test_desc",
				}, nil)
				checksRepo.EXPECT().Exists(gomock.Any(), habitID, checkDate).Return(true, nil)
			},
		},
		{
			Desc:      "error habit not found",
			Error:     errorvalues.ErrHabitNotFound,
			HabitID:   habitID,
			UserID:    userID,
			CheckDate: checkDate,
			MockPrepFunc: func() {
				habitsRepo.EXPECT().GetByID(gomock.Any(), habitID).Return(nil, errorvalues.ErrHabitNotFound)
			},
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.Desc, func(t *testing.T) {
			tc.MockPrepFunc()
			err := serv.CheckHabit(ctx, tc.HabitID, tc.UserID, tc.CheckDate)
			assert.ErrorIs(t, err, tc.Error)
		})
	}
}

func TestUncheckHabit(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	checksRepo := mocks.NewMockHabitChecksRepositoryI(ctrl)
	habitsRepo := mocks.NewMockHabitsRepositoryI(ctrl)

	serv := service.NewHabitChecksService(habitsRepo, checksRepo)
	habitID := uuid.New()
	userID := uuid.New()
	checkDate := time.Now()
	testCases := []struct {
		Desc         string
		Error        error
		HabitID      uuid.UUID
		UserID       uuid.UUID
		CheckDate    time.Time
		MockPrepFunc func()
	}{
		{
			Desc:      "success",
			Error:     nil,
			HabitID:   habitID,
			UserID:    userID,
			CheckDate: checkDate,
			MockPrepFunc: func() {
				habitsRepo.EXPECT().GetByID(gomock.Any(), habitID).Return(&entity.Habit{
					ID:          habitID,
					UserID:      userID,
					Title:       "test_habit",
					Description: "test_desc",
				}, nil)
				checksRepo.EXPECT().Exists(gomock.Any(), habitID, checkDate).Return(true, nil)
				checksRepo.EXPECT().Delete(gomock.Any(), habitID, checkDate).Return(nil)
			},
		},
		{
			Desc:      "error wrong owner",
			Error:     errorvalues.ErrWrongOwner,
			HabitID:   habitID,
			UserID:    userID,
			CheckDate: checkDate,
			MockPrepFunc: func() {
				habitsRepo.EXPECT().GetByID(gomock.Any(), habitID).Return(&entity.Habit{
					ID:          habitID,
					UserID:      uuid.New(),
					Title:       "test_habit",
					Description: "test_desc",
				}, nil)
			},
		},
		{
			Desc:      "error deleted unexisted check",
			Error:     errorvalues.ErrCheckNotFound,
			HabitID:   habitID,
			UserID:    userID,
			CheckDate: checkDate,
			MockPrepFunc: func() {
				habitsRepo.EXPECT().GetByID(gomock.Any(), habitID).Return(&entity.Habit{
					ID:          habitID,
					UserID:      userID,
					Title:       "test_habit",
					Description: "test_desc",
				}, nil)
				checksRepo.EXPECT().Exists(gomock.Any(), habitID, checkDate).Return(false, nil)
			},
		},
		{
			Desc:      "error habit not found",
			Error:     errorvalues.ErrHabitNotFound,
			HabitID:   habitID,
			UserID:    userID,
			CheckDate: checkDate,
			MockPrepFunc: func() {
				habitsRepo.EXPECT().GetByID(gomock.Any(), habitID).Return(nil, errorvalues.ErrHabitNotFound)
			},
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.Desc, func(t *testing.T) {
			tc.MockPrepFunc()
			err := serv.UncheckHabit(ctx, tc.HabitID, tc.UserID, tc.CheckDate)
			assert.ErrorIs(t, err, tc.Error)
		})
	}
}

func TestGetHabitChecks(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	checksRepo := mocks.NewMockHabitChecksRepositoryI(ctrl)
	habitsRepo := mocks.NewMockHabitsRepositoryI(ctrl)

	serv := service.NewHabitChecksService(habitsRepo, checksRepo)
	habitID := uuid.New()
	userID := uuid.New()
	returnedChecks := make([]entity.HabitCheck, 0, 5)
	now := time.Now()
	now = now.Truncate(24 * time.Hour)
	for i := range cap(returnedChecks) {
		returnedChecks = append(returnedChecks, entity.HabitCheck{
			ID:        i + 1,
			HabitID:   habitID,
			CheckDate: now.Add(time.Hour * 24 * time.Duration(-i)),
			CreatedAt: now.Add(time.Hour * 24 * time.Duration(-i)),
		})
	}
	// Counting from 4 days earlier
	from := returnedChecks[len(returnedChecks)-1].CheckDate
	testCases := []struct {
		Desc      string
		Error     error
		HabitID   uuid.UUID
		UserID    uuid.UUID
		DateRange struct {
			From time.Time
			To   time.Time
		}
		Result       []entity.HabitCheck
		MockPrepFunc func()
	}{
		{
			Desc:    "success",
			Error:   nil,
			HabitID: habitID,
			UserID:  userID,
			Result:  returnedChecks,
			DateRange: struct {
				From time.Time
				To   time.Time
			}{
				From: from,
				To:   now,
			},
			MockPrepFunc: func() {
				habitsRepo.EXPECT().GetByID(gomock.Any(), habitID).Return(&entity.Habit{
					ID:          habitID,
					UserID:      userID,
					Title:       "test_habit",
					Description: "test_desc",
				}, nil)
				checksRepo.EXPECT().
					GetByHabitAndDateRange(gomock.Any(), habitID, from, now).
					Return(returnedChecks, nil)
			},
		},
		{
			Desc:    "error wrong owner",
			Error:   errorvalues.ErrWrongOwner,
			HabitID: habitID,
			UserID:  userID,
			Result:  nil,
			DateRange: struct {
				From time.Time
				To   time.Time
			}{
				From: from,
				To:   now,
			},
			MockPrepFunc: func() {
				habitsRepo.EXPECT().GetByID(gomock.Any(), habitID).Return(&entity.Habit{
					ID:          habitID,
					UserID:      uuid.New(),
					Title:       "test_habit",
					Description: "test_desc",
				}, nil)
			},
		},
		{
			Desc:    "error habit not found",
			Error:   errorvalues.ErrHabitNotFound,
			HabitID: habitID,
			UserID:  userID,
			Result:  nil,
			DateRange: struct {
				From time.Time
				To   time.Time
			}{
				From: from,
				To:   now,
			},
			MockPrepFunc: func() {
				habitsRepo.EXPECT().GetByID(gomock.Any(), habitID).Return(nil, errorvalues.ErrHabitNotFound)
			},
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.Desc, func(t *testing.T) {
			tc.MockPrepFunc()
			result, err := serv.GetHabitChecks(ctx, tc.HabitID, tc.UserID, tc.DateRange.From, tc.DateRange.To)
			assert.ErrorIs(t, err, tc.Error)
			assert.Equal(t, tc.Result, result)
		})
	}
}

func TestGetHabitStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	habitsRepo := mocks.NewMockHabitsRepositoryI(ctrl)
	checksRepo := mocks.NewMockHabitChecksRepositoryI(ctrl)

	svc := service.NewHabitChecksService(habitsRepo, checksRepo)
	ctx := context.Background()

	habit := entity.Habit{
		ID:          habitID,
		UserID:      userID,
		Title:       "brush teeth",
		Description: "get your teeth clean",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	habitStats := entity.HabitStats{
		ID:            habit.ID,
		TotalChecks:   10,
		CurrentStreak: 5,
		MaxStreak:     5,
		LastCheck:     time.Now(),
	}

	testCases := []struct {
		caseName     string
		habitID      uuid.UUID
		userID       uuid.UUID
		result       *entity.HabitStats
		mockPrepFunc func()
		Error        error
	}{
		{
			caseName: "success",
			habitID:  habit.ID,
			userID:   habit.UserID,
			result:   &habitStats,
			mockPrepFunc: func() {
				habitsRepo.EXPECT().
					GetByID(gomock.Any(), habit.ID).
					Return(&habit, nil)
				checksRepo.EXPECT().
					CountByHabitID(gomock.Any(), habit.ID).
					Return(10, nil)
				checksRepo.EXPECT().
					GetCurrentStreak(gomock.Any(), habit.ID).
					Return(5, nil)
				checksRepo.EXPECT().
					GetMaxStreak(gomock.Any(), habit.ID).
					Return(5, nil)
				checksRepo.EXPECT().
					GetLastCheckDate(gomock.Any(), habit.ID).
					Return(&habitStats.LastCheck, nil)
			},
			Error: nil,
		},
		{
			caseName: "error: habit not found",
			habitID:  habit.ID,
			userID:   habit.UserID,
			mockPrepFunc: func() {
				habitsRepo.EXPECT().
					GetByID(gomock.Any(), habit.ID).
					Return(nil, errorvalues.ErrHabitNotFound)
			},
			Error: errorvalues.ErrHabitNotFound,
		},
		{
			caseName: "error: wrong owner",
			habitID:  habit.ID,
			userID:   uuid.New(),
			mockPrepFunc: func() {
				habitsRepo.EXPECT().
					GetByID(gomock.Any(), habit.ID).
					Return(&habit, nil)
			},
			Error: errorvalues.ErrWrongOwner,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.mockPrepFunc()
			stats, err := svc.GetHabitStats(ctx, tc.habitID, tc.userID)
			assert.ErrorIs(t, err, tc.Error)
			if err == nil && tc.Error == nil {
				assert.Equal(t, *tc.result, *stats)
			}
		})
	}
}

func TestHabitChecksServiceIntegral(t *testing.T) {
	cfg := setupHabitsTestDB(t)
	habitRepo := repository.NewHabitsRepo(cfg)
	checksRepo := repository.NewHabitChecksRepo(cfg)

	ctx := context.Background()
	var (
		habit *entity.Habit
		err   error
	)
	svc := service.NewHabitChecksService(habitRepo, checksRepo)
	// Creating a habit to operate
	{
		svc := service.NewHabitsService(habitRepo)
		habit, err = svc.CreateHabit(ctx, userID, service.CreateHabitRequest{
			Title:       "brush teeth",
			Description: "get your teeth clean",
		})
		require.NoError(t, err)
	}

	date := time.Now()

	t.Run("checking", func(t *testing.T) {
		testCases := []struct {
			caseName string
			habitID  uuid.UUID
			userID   uuid.UUID
			date     time.Time
			Error    error
		}{
			{
				caseName: "success",
				habitID:  habit.ID,
				userID:   userID,
				date:     date,
				Error:    nil,
			},
			{
				caseName: "error: habit not found",
				habitID:  uuid.New(),
				userID:   userID,
				date:     date,
				Error:    errorvalues.ErrHabitNotFound,
			},
			{
				caseName: "error: wrong owner",
				habitID:  habit.ID,
				userID:   uuid.New(),
				date:     date,
				Error:    errorvalues.ErrWrongOwner,
			},
		}
		for _, tc := range testCases {
			err = svc.CheckHabit(ctx, tc.habitID, tc.userID, tc.date)
			assert.ErrorIs(t, err, tc.Error)
		}
		for i := 1; i <= 10; i++ {
			if i == 3 {
				continue
			}
			err = svc.CheckHabit(ctx, habit.ID, userID, date.Add(-time.Duration(i)*time.Hour*24))
		}
	})
	t.Run("unchecking", func(t *testing.T) {
		testCases := []struct {
			caseName string
			habitID  uuid.UUID
			userID   uuid.UUID
			date     time.Time
			Error    error
		}{
			{
				caseName: "success",
				habitID:  habit.ID,
				userID:   userID,
				date:     date,
				Error:    nil,
			},
			{
				caseName: "error: habit not found",
				habitID:  uuid.New(),
				userID:   userID,
				date:     date,
				Error:    errorvalues.ErrHabitNotFound,
			},
			{
				caseName: "error: wrong owner",
				habitID:  habit.ID,
				userID:   uuid.New(),
				date:     date,
				Error:    errorvalues.ErrWrongOwner,
			},
		}
		for _, tc := range testCases {
			err = svc.UncheckHabit(ctx, tc.habitID, tc.userID, tc.date)
			assert.ErrorIs(t, err, tc.Error)
		}
	})
	t.Run("get checks", func(t *testing.T) {
		checks, err := svc.GetHabitChecks(ctx, habit.ID, userID, date.Add(-10*time.Hour*24), date)
		assert.NoError(t, err)
		assert.Equal(t, 9, len(checks))
	})
	t.Run("getting stats", func(t *testing.T) {
		stats, err := svc.GetHabitStats(ctx, habit.ID, userID)
		assert.NoError(t, err)
		assert.Equal(t, entity.HabitStats{
			ID:            habit.ID,
			TotalChecks:   9,
			CurrentStreak: 0,
			MaxStreak:     7,
			LastCheck:     stats.LastCheck,
		}, *stats)
	})
}
