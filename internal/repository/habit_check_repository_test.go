package repository_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/pkg/entity"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCheck(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	habitChecksRepo := repository.NewHabitChecksRepoWithConn(mock)
	query := regexp.QuoteMeta(`INSERT INTO habit_checks (habit_id, check_date) VALUES ($1, $2);`)
	habitID := uuid.New()
	checkDate := time.Now()
	testCases := []struct {
		Desc            string
		Error           error
		MockPrepareFunc func()
	}{
		{
			Desc:  "successful",
			Error: nil,
			MockPrepareFunc: func() {
				mock.ExpectExec(query).WithArgs(habitID, checkDate).WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			Desc:  "unique violation",
			Error: errorvalues.ErrCheckExist,
			MockPrepareFunc: func() {
				mock.ExpectExec(query).WithArgs(habitID, checkDate).WillReturnError(&pgconn.PgError{
					Code: "23505",
				})
			},
		},
		{
			Desc:  "fk violation",
			Error: errorvalues.ErrHabitNotFound,
			MockPrepareFunc: func() {
				mock.ExpectExec(query).WithArgs(habitID, checkDate).WillReturnError(&pgconn.PgError{
					Code: "23503",
				})
			},
		},
		{
			Desc:  "db error",
			Error: errors.New("creating check error: db error"),
			MockPrepareFunc: func() {
				mock.ExpectExec(query).WithArgs(habitID, checkDate).WillReturnError(errors.New("db error"))
			},
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.Desc, func(t *testing.T) {
			tc.MockPrepareFunc()
			err := habitChecksRepo.Create(ctx, habitID, checkDate)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteCheck(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	habitChecksRepo := repository.NewHabitChecksRepoWithConn(mock)
	query := regexp.QuoteMeta(`DELETE FROM habit_checks WHERE habit_id = $1 AND check_date = $2;`)
	habitID := uuid.New()
	checkDate := time.Now()
	testCases := []struct {
		Desc         string
		Error        error
		MockPrepFunc func()
	}{
		{
			Desc:  "successful",
			Error: nil,
			MockPrepFunc: func() {
				mock.ExpectExec(query).WithArgs(habitID, checkDate).WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
		},
		{
			Desc:  "db error",
			Error: errors.New("deleting check error: db error"),
			MockPrepFunc: func() {
				mock.ExpectExec(query).WithArgs(habitID, checkDate).WillReturnError(errors.New("db error"))
			},
		},
		{
			Desc:  "check not found",
			Error: errorvalues.ErrCheckNotFound,
			MockPrepFunc: func() {
				mock.ExpectExec(query).WithArgs(habitID, checkDate).WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.Desc, func(t *testing.T) {
			tc.MockPrepFunc()
			err := habitChecksRepo.Delete(ctx, habitID, checkDate)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})

	}
}

func TestExistsCheck(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	habitChecksRepo := repository.NewHabitChecksRepoWithConn(mock)
	query := regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM habit_checks WHERE habitID = $1 AND check_date = $2);`)
	habitID := uuid.New()
	checkDate := time.Now()
	testCases := []struct {
		Desc          string
		Error         error
		IsExistResult bool
		MockPrepFunc  func()
	}{
		{
			Desc:  "successful: exists",
			Error: nil,
			MockPrepFunc: func() {
				mock.ExpectQuery(query).
					WithArgs(habitID, checkDate).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
			},
			IsExistResult: true,
		},
		{
			Desc:  "successful: doesn't exist",
			Error: nil,
			MockPrepFunc: func() {
				mock.ExpectQuery(query).
					WithArgs(habitID, checkDate).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))
			},
			IsExistResult: false,
		},
		{
			Desc:  "db error",
			Error: errors.New("inspecting if check exists error: db error"),
			MockPrepFunc: func() {
				mock.ExpectQuery(query).
					WithArgs(habitID, checkDate).
					WillReturnError(errors.New("db error"))
			},
			IsExistResult: false,
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.Desc, func(t *testing.T) {
			tc.MockPrepFunc()
			exists, err := habitChecksRepo.Exists(ctx, habitID, checkDate)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.IsExistResult, exists)
			}
		})
	}
}

func TestGetByHabitAndDateRange(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	habitChecksRepo := repository.NewHabitChecksRepoWithConn(mock)
	query := regexp.QuoteMeta(`SELECT id, habit_id, check_date, created_at FROM habit_checks WHERE habitID = $1 AND check_date >= $2 AND check_date <= $3;`)
	habitID := uuid.New()
	fromDate := time.Now().Add(time.Hour * -24)
	toDate := time.Now().Add(time.Hour * 24)
	returnedChecks := []entity.HabitCheck{
		{
			ID:        1,
			HabitID:   habitID,
			CheckDate: fromDate,
			CreatedAt: fromDate,
		},
		{
			ID:        2,
			HabitID:   habitID,
			CheckDate: time.Now(),
			CreatedAt: time.Now(),
		},
		{
			ID:        3,
			HabitID:   habitID,
			CheckDate: toDate,
			CreatedAt: toDate,
		},
	}
	testCases := []struct {
		Desc         string
		Error        error
		ChecksResult []entity.HabitCheck
		MockPrepFunc func()
	}{
		{
			Desc:         "success",
			Error:        nil,
			ChecksResult: returnedChecks,
			MockPrepFunc: func() {
				rows := pgxmock.NewRows([]string{"id", "habit_id", "check_date", "created_at"})
				for _, check := range returnedChecks {
					rows.AddRow(check.ID, check.HabitID, check.CheckDate, check.CreatedAt)
				}
				mock.ExpectQuery(query).
					WithArgs(habitID, fromDate, toDate).
					WillReturnRows(rows)
			},
		},
		{
			Desc:         "db error",
			Error:        errors.New("getting checks for period error: db error"),
			ChecksResult: nil,
			MockPrepFunc: func() {
				mock.ExpectQuery(query).
					WithArgs(habitID, fromDate, toDate).
					WillReturnError(errors.New("db error"))
			},
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.Desc, func(t *testing.T) {
			tc.MockPrepFunc()
			result, err := habitChecksRepo.GetByHabitAndDateRange(ctx, habitID, fromDate, toDate)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.ChecksResult, result)
			}
		})
	}
}

func TestGetLastCheckDate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	habitChecksRepo := repository.NewHabitChecksRepoWithConn(mock)
	query := regexp.QuoteMeta(`SELECT check_date FROM habit_checks WHERE habit_id = $1 ORDER BY check_date DESC LIMIT 1;`)
	habitID := uuid.New()
	returnedDate := time.Now().Add(time.Hour * -24)
	testCases := []struct {
		Desc            string
		Error           error
		CheckDateResult *time.Time
		MockPrepFunc    func()
	}{
		{
			Desc:            "successful",
			Error:           nil,
			CheckDateResult: &returnedDate,
			MockPrepFunc: func() {
				mock.ExpectQuery(query).
					WithArgs(habitID).
					WillReturnRows(pgxmock.NewRows([]string{"check_date"}).AddRow(returnedDate))
			},
		},
		{
			Desc:            "ErrNoRows",
			Error:           nil,
			CheckDateResult: nil,
			MockPrepFunc: func() {
				mock.ExpectQuery(query).
					WithArgs(habitID).
					WillReturnError(pgx.ErrNoRows)
			},
		},
		{
			Desc:            "other db error",
			Error:           errors.New("getting last check date error: db error"),
			CheckDateResult: nil,
			MockPrepFunc: func() {
				mock.ExpectQuery(query).
					WithArgs(habitID).
					WillReturnError(errors.New("db error"))
			},
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.Desc, func(t *testing.T) {
			tc.MockPrepFunc()
			date, err := habitChecksRepo.GetLastCheckDate(ctx, habitID)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
			if tc.CheckDateResult == nil {
				assert.Nil(t, date)
			} else {
				require.NotNil(t, date)
				assert.Equal(t, *tc.CheckDateResult, *date)
			}
		})
	}
}

func TestCountByHabitID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	habitChecksRepo := repository.NewHabitChecksRepoWithConn(mock)
	query := regexp.QuoteMeta(`SELECT COUNT(*) FROM habit_checks WHERE habit_id = $1;`)
	habitID := uuid.New()
	testCases := []struct {
		Desc         string
		Error        error
		CountResult  int
		MockPrepFunc func()
	}{
		{
			Desc:        "successful",
			Error:       nil,
			CountResult: 10,
			MockPrepFunc: func() {
				mock.ExpectQuery(query).
					WithArgs(habitID).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(10))
			},
		},
		{
			Desc:  "db error",
			Error: errors.New("error counting checks: db error"),
			MockPrepFunc: func() {
				mock.ExpectQuery(query).
					WithArgs(habitID).
					WillReturnError(errors.New("db error"))
			},
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.Desc, func(t *testing.T) {
			tc.MockPrepFunc()
			count, err := habitChecksRepo.CountByHabitID(ctx, habitID)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, count, tc.CountResult)
			}
		})
	}
}
