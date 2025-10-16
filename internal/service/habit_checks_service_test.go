package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/repository/mocks"
	"github.com/limbo/discipline/internal/service"
	"github.com/limbo/discipline/pkg/entity"
	"github.com/stretchr/testify/assert"
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
