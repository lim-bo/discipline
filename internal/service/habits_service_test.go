package service_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/internal/service"
	"github.com/limbo/discipline/pkg/entity"
	"github.com/pressly/goose"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type mockState int

const (
	stateSuccess = iota
	stateDBError
	stateUserHasHabitError
	stateHabitNotFoundError
	stateUserNotFoundError
	stateWrongOwner
)

type habitRepoMock struct {
	state mockState
}

// Variables for tests
var (
	userID       = uuid.New()
	userName     = "test_owner"
	userPassHash = "test_passhash"
	habitID      = uuid.New()
	testHabit    = entity.Habit{
		ID:          habitID,
		UserID:      userID,
		Title:       "test_habit",
		Description: "test_description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
)

func (hrmock *habitRepoMock) Create(ctx context.Context, habit *entity.Habit) (uuid.UUID, error) {
	switch hrmock.state {
	case stateUserNotFoundError:
		return uuid.UUID{}, errorvalues.ErrOwnerNotFound
	case stateUserHasHabitError:
		return uuid.UUID{}, errorvalues.ErrUserHasHabit
	case stateDBError:
		return uuid.UUID{}, errors.New("db error")
	default:
		return habitID, nil
	}
}

func (hrmock *habitRepoMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Habit, error) {
	switch hrmock.state {
	case stateHabitNotFoundError:
		return nil, errorvalues.ErrHabitNotFound
	case stateDBError:
		return nil, errors.New("db error")
	case stateWrongOwner:
		return &entity.Habit{
			ID:          testHabit.ID,
			UserID:      uuid.New(),
			Title:       testHabit.Title,
			Description: testHabit.Description,
			CreatedAt:   testHabit.CreatedAt,
			UpdatedAt:   testHabit.UpdatedAt,
		}, nil
	default:
		return &testHabit, nil
	}
}

func (hrmock *habitRepoMock) GetByUserID(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*entity.Habit, error) {
	switch hrmock.state {
	case stateUserNotFoundError:
		return []*entity.Habit{}, nil
	case stateDBError:
		return nil, errors.New("db error")
	default:
		return []*entity.Habit{
			&testHabit,
		}, nil
	}
}
func (hrmock *habitRepoMock) Update(ctx context.Context, habit *entity.Habit) error {
	switch hrmock.state {
	case stateDBError:
		return errors.New("db error")
	case stateHabitNotFoundError:
		return errorvalues.ErrHabitNotFound
	default:
		return nil
	}
}
func (hrmock *habitRepoMock) Delete(ctx context.Context, id uuid.UUID) error {
	switch hrmock.state {
	case stateDBError:
		return errors.New("db error")
	case stateHabitNotFoundError:
		return errorvalues.ErrHabitNotFound
	default:
		return nil
	}
}

func TestCreateHabit(t *testing.T) {
	mock := &habitRepoMock{state: stateSuccess}
	s := service.NewHabitsService(mock)
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		h, err := s.CreateHabit(ctx, userID, service.CreateHabitRequest{
			Title:       testHabit.Title,
			Description: testHabit.Description,
		})
		assert.NoError(t, err)
		assert.Equal(t, testHabit, *h)
	})
	t.Run("db error", func(t *testing.T) {
		mock.state = stateDBError
		_, err := s.CreateHabit(ctx, userID, service.CreateHabitRequest{
			Title:       testHabit.Title,
			Description: testHabit.Description,
		})
		assert.Error(t, err)
	})
	t.Run("owner not found", func(t *testing.T) {
		mock.state = stateUserNotFoundError
		_, err := s.CreateHabit(ctx, userID, service.CreateHabitRequest{
			Title:       testHabit.Title,
			Description: testHabit.Description,
		})
		assert.ErrorIs(t, err, errorvalues.ErrUserNotFound)
	})
	t.Run("habit duplication", func(t *testing.T) {
		mock.state = stateUserHasHabitError
		_, err := s.CreateHabit(ctx, userID, service.CreateHabitRequest{
			Title:       testHabit.Title,
			Description: testHabit.Description,
		})
		assert.ErrorIs(t, err, errorvalues.ErrUserHasHabit)
	})
}

func TestGetUserHabits(t *testing.T) {
	mock := &habitRepoMock{state: stateSuccess}
	s := service.NewHabitsService(mock)
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		habits, err := s.GetUserHabits(
			ctx,
			userID,
			service.PaginationOpts{
				Limit:  10,
				Offset: 0,
			},
		)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(habits))
		assert.Equal(t, testHabit, *habits[0])
	})
	t.Run("db error", func(t *testing.T) {
		mock.state = stateDBError
		_, err := s.GetUserHabits(
			ctx,
			userID,
			service.PaginationOpts{
				Limit:  10,
				Offset: 0,
			},
		)
		assert.Error(t, err)
	})
}

func TestGetHabitByID(t *testing.T) {
	mock := &habitRepoMock{state: stateSuccess}
	s := service.NewHabitsService(mock)
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		h, err := s.GetHabit(ctx, habitID, userID)
		assert.NoError(t, err)
		assert.Equal(t, testHabit, *h)
	})
	t.Run("wrong owner", func(t *testing.T) {
		mock.state = stateWrongOwner
		_, err := s.GetHabit(ctx, habitID, userID)
		assert.ErrorIs(t, err, errorvalues.ErrWrongOwner)
	})
	t.Run("db error", func(t *testing.T) {
		mock.state = stateDBError
		_, err := s.GetHabit(ctx, habitID, userID)
		assert.Error(t, err)
	})
	t.Run("habit not found", func(t *testing.T) {
		mock.state = stateHabitNotFoundError
		_, err := s.GetHabit(ctx, habitID, userID)
		assert.ErrorIs(t, err, errorvalues.ErrHabitNotFound)
	})
}

func TestDeleteHabit(t *testing.T) {
	mock := &habitRepoMock{state: stateSuccess}
	s := service.NewHabitsService(mock)
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		err := s.DeleteHabit(ctx, habitID, userID)
		assert.NoError(t, err)
	})
	t.Run("wrong owner", func(t *testing.T) {
		mock.state = stateWrongOwner
		err := s.DeleteHabit(ctx, habitID, userID)
		assert.ErrorIs(t, err, errorvalues.ErrWrongOwner)
	})
	t.Run("habit not found", func(t *testing.T) {
		mock.state = stateHabitNotFoundError
		err := s.DeleteHabit(ctx, habitID, userID)
		assert.ErrorIs(t, err, errorvalues.ErrHabitNotFound)
	})
	t.Run("db error", func(t *testing.T) {
		mock.state = stateDBError
		err := s.DeleteHabit(ctx, habitID, userID)
		assert.Error(t, err)
	})
}

func TestHabitsServiceIntegrational(t *testing.T) {
	cfg := setupHabitsTestDB(t)
	repo := repository.NewHabitsRepo(cfg)
	s := service.NewHabitsService(repo)
	habits := []*entity.Habit{}
	for i := range 5 {
		habits = append(habits, &entity.Habit{
			Title:       fmt.Sprintf("test_habit_%d", i),
			Description: fmt.Sprintf("test_description_%d", i),
		})
	}
	ctx := context.Background()
	t.Run("create habit", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			for i, h := range habits {
				res, err := s.CreateHabit(ctx, userID, service.CreateHabitRequest{
					Title:       h.Title,
					Description: h.Description,
				})
				assert.NoError(t, err)
				assert.Equal(t, res.Title, h.Title)
				assert.Equal(t, res.Description, h.Description)
				habits[i] = res
			}
		})
		t.Run("error: unexist user", func(t *testing.T) {
			_, err := s.CreateHabit(ctx, uuid.New(), service.CreateHabitRequest{
				Title:       "aaa",
				Description: "bbb",
			})
			assert.ErrorIs(t, err, errorvalues.ErrUserNotFound)
		})
		t.Run("error: habit exists", func(t *testing.T) {
			_, err := s.CreateHabit(ctx, userID, service.CreateHabitRequest{
				Title:       habits[0].Title,
				Description: habits[0].Description,
			})
			assert.ErrorIs(t, err, errorvalues.ErrUserHasHabit)
		})
	})
	t.Run("get user's habits", func(t *testing.T) {
		t.Run("got all", func(t *testing.T) {
			result, err := s.GetUserHabits(ctx, userID, service.PaginationOpts{Limit: 5, Offset: 0})
			assert.NoError(t, err)
			assert.Equal(t, 5, len(result))
			for i := range result {
				assert.Equal(t, *habits[i], *result[i])
			}
		})
		t.Run("got some", func(t *testing.T) {
			limit, offset := 2, 2
			result, err := s.GetUserHabits(ctx, userID, service.PaginationOpts{Limit: limit, Offset: offset})
			assert.NoError(t, err)
			assert.Equal(t, limit, len(result))
			for i := range limit {
				assert.Equal(t, *habits[i+offset], *result[i])
			}
		})
	})

	t.Run("get habit by id", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			h, err := s.GetHabit(ctx, habits[0].ID, userID)
			assert.NoError(t, err)
			assert.Equal(t, *habits[0], *h)
		})
		t.Run("error: wrong owner", func(t *testing.T) {
			_, err := s.GetHabit(ctx, habits[0].ID, uuid.New())
			assert.ErrorIs(t, err, errorvalues.ErrWrongOwner)
		})
		t.Run("error: habit not found", func(t *testing.T) {
			_, err := s.GetHabit(ctx, uuid.New(), userID)
			assert.ErrorIs(t, err, errorvalues.ErrHabitNotFound)
		})
	})

	t.Run("delete habit", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			err := s.DeleteHabit(ctx, habits[0].ID, userID)
			assert.NoError(t, err)
		})
		t.Run("error: wrong owner", func(t *testing.T) {
			err := s.DeleteHabit(ctx, habits[1].ID, uuid.New())
			assert.ErrorIs(t, err, errorvalues.ErrWrongOwner)
		})
		t.Run("error: habit not found", func(t *testing.T) {
			err := s.DeleteHabit(ctx, habits[0].ID, userID)
			assert.ErrorIs(t, err, errorvalues.ErrHabitNotFound)
		})
	})
}

func setupHabitsTestDB(t *testing.T) *testPGConfig {
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
	_, err = conn.Exec(`INSERT INTO users (id, name, password_hash) VALUES ($1, $2, $3);`, userID, userName, userPassHash)
	if err != nil {
		t.Fatal("adding mock user error: " + err.Error())
	}
	conn.Close()
	t.Cleanup(func() {
		container.Terminate(context.Background())
	})
	return &testPGConfig{
		connStr: connStr,
	}
}
