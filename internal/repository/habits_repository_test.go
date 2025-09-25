package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/lib/pq"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/pkg/entity"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/pressly/goose"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	userID = uuid.New()
)

func TestCreateHabit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewHabitsRepoWithConn(mock)
	habit := entity.Habit{
		UserID:      userID,
		Title:       "test_habit",
		Description: "blah blah blah",
	}
	hid := uuid.New()
	ctx := context.Background()
	query := regexp.QuoteMeta(`INSERT INTO habits (user_id, title, description) VALUES ($1, $2, $3);`)
	selectQuery := regexp.QuoteMeta(`SELECT id FROM habits WHERE title = $1 AND user_id = $2;`)
	t.Run("successfully created", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(query).
			WithArgs(habit.UserID, habit.Title, habit.Description).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		mock.ExpectQuery(selectQuery).
			WithArgs(habit.Title, habit.UserID).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(hid))
		mock.ExpectCommit()
		id, err := repo.Create(ctx, &habit)
		assert.NoError(t, err)
		assert.Equal(t, hid, id)
	})
	t.Run("Unique violation", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(query).
			WithArgs(habit.UserID, habit.Title, habit.Description).
			WillReturnError(&pgconn.PgError{Code: "23505"})
		mock.ExpectRollback()
		_, err := repo.Create(ctx, &habit)
		assert.ErrorIs(t, err, errorvalues.ErrUserHasHabit)
	})
	t.Run("FK violation", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(query).
			WithArgs(habit.UserID, habit.Title, habit.Description).
			WillReturnError(&pgconn.PgError{Code: "23503"})
		mock.ExpectRollback()
		_, err := repo.Create(ctx, &habit)
		assert.ErrorIs(t, err, errorvalues.ErrOwnerNotFound)
	})
	t.Run("db error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(query).
			WithArgs(habit.UserID, habit.Title, habit.Description).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()
		_, err := repo.Create(ctx, &habit)
		assert.Error(t, err)
	})
}

func TestGetHabitByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewHabitsRepoWithConn(mock)
	habit := entity.Habit{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       "test_habit",
		Description: "blah blah blah",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	query := regexp.QuoteMeta(`SELECT user_id, title, description, created_at, updated_at FROM habits WHERE id = $1;`)
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(query).
			WithArgs(habit.ID).
			WillReturnRows(pgxmock.NewRows([]string{"user_id", "title", "description", "created_at", "updated_at"}).
				AddRow(habit.UserID, habit.Title, habit.Description, habit.CreatedAt, habit.UpdatedAt),
			)
		result, err := repo.GetByID(ctx, habit.ID)
		assert.NoError(t, err)
		assert.Equal(t, habit, *result)
	})
	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(query).
			WithArgs(habit.ID).
			WillReturnError(pgx.ErrNoRows)
		_, err := repo.GetByID(ctx, habit.ID)
		assert.ErrorIs(t, err, errorvalues.ErrHabitNotFound)
	})
	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(query).
			WithArgs(habit.ID).
			WillReturnError(errors.New("db error"))
		_, err := repo.GetByID(ctx, habit.ID)
		assert.Error(t, err)
	})
}

func TestGetHabitsByUserID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewHabitsRepoWithConn(mock)
	habits := []*entity.Habit{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Title:     "test_habit_1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			Title:     "test_habit_2",
			CreatedAt: time.Now().Add(time.Hour),
			UpdatedAt: time.Now().Add(time.Hour),
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			Title:     "test_habit_3",
			CreatedAt: time.Now().Add(time.Hour * 2),
			UpdatedAt: time.Now().Add(time.Hour * 2),
		},
	}
	query := regexp.QuoteMeta(`SELECT id, user_id, title, description, created_at, updated_at 
		FROM habits WHERE user_id = $1 LIMIT $2 OFFSET $3;`)
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		limit := 3
		offset := 0
		rows := pgxmock.NewRows([]string{"id", "user_id", "title", "description", "created_at", "updated_at"})
		for _, h := range habits {
			rows.AddRow(h.ID, h.UserID, h.Title, h.Description, h.CreatedAt, h.UpdatedAt)
		}
		mock.ExpectQuery(query).
			WithArgs(userID, limit, offset).
			WillReturnRows(rows)
		result, err := repo.GetByUserID(ctx, userID, limit, offset)
		assert.NoError(t, err)
		for i := range result {
			assert.Equal(t, *habits[i], *result[i])
		}
	})
	t.Run("used limit and offset", func(t *testing.T) {
		limit := 1
		offset := 1
		rows := pgxmock.NewRows([]string{"id", "user_id", "title", "description", "created_at", "updated_at"})
		rows.AddRow(habits[1].ID, habits[1].UserID, habits[1].Title, habits[1].Description, habits[1].CreatedAt, habits[1].UpdatedAt)
		mock.ExpectQuery(query).
			WithArgs(userID, limit, offset).
			WillReturnRows(rows)
		result, err := repo.GetByUserID(ctx, userID, limit, offset)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, *habits[1], *result[0])
	})
	t.Run("db error", func(t *testing.T) {
		limit := 1
		offset := 1
		mock.ExpectQuery(query).
			WithArgs(userID, limit, offset).
			WillReturnError(errors.New("db error"))
		_, err := repo.GetByUserID(ctx, userID, limit, offset)
		assert.Error(t, err)
	})
}

func TestUpdateHabit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewHabitsRepoWithConn(mock)
	query := regexp.QuoteMeta(`UPDATE habits SET title = $1, description = $2, updated_at = NOW() WHERE id = $3;`)
	habit := entity.Habit{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       "test_habit",
		Description: "blah blah blah",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(query).
			WithArgs(habit.Title, habit.Description, habit.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		err := repo.Update(ctx, &habit)
		assert.NoError(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(query).
			WithArgs(habit.Title, habit.Description, habit.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))
		err := repo.Update(ctx, &habit)
		assert.ErrorIs(t, err, errorvalues.ErrHabitNotFound)
	})
	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(query).
			WithArgs(habit.Title, habit.Description, habit.ID).
			WillReturnError(errors.New("db error"))
		err := repo.Update(ctx, &habit)
		assert.Error(t, err)
	})
}

func TestDeleteHabit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewHabitsRepoWithConn(mock)
	query := regexp.QuoteMeta(`DELETE FROM habits WHERE id = $1;`)
	ctx := context.Background()
	id := uuid.New()
	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(query).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(query).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))
		err := repo.Delete(ctx, id)
		assert.ErrorIs(t, err, errorvalues.ErrHabitNotFound)
	})
	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(query).
			WithArgs(id).
			WillReturnError(errors.New("db error"))
		err := repo.Delete(ctx, id)
		assert.Error(t, err)
	})
}

func TestHabitsIntegrational(t *testing.T) {
	cfg := setupHabitsTestDB(t)
	repo := repository.NewHabitsRepo(cfg)
	habits := []*entity.Habit{}
	for i := range 5 {
		habits = append(habits, &entity.Habit{
			UserID:      userID,
			Title:       fmt.Sprintf("habit_n%d", i),
			Description: fmt.Sprintf("desc_n%d", i),
		})
	}
	ctx := context.Background()
	t.Run("create", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			id, err := repo.Create(ctx, habits[0])
			assert.NoError(t, err)
			habits[0].ID = id
		})
		t.Run("already exist error", func(t *testing.T) {
			_, err := repo.Create(ctx, habits[0])
			assert.ErrorIs(t, err, errorvalues.ErrUserHasHabit)
		})
		t.Run("unknown user error", func(t *testing.T) {
			_, err := repo.Create(ctx, &entity.Habit{
				UserID:      uuid.New(),
				Title:       "ttt",
				Description: "ddd",
			})
			assert.ErrorIs(t, err, errorvalues.ErrOwnerNotFound)
		})
		t.Run("append more", func(t *testing.T) {
			for i := 1; i < 5; i++ {
				id, err := repo.Create(ctx, habits[i])
				assert.NoError(t, err)
				habits[i].ID = id
				t.Log(id)
			}
		})
	})
	t.Run("get habits by user_id", func(t *testing.T) {
		t.Run("list all habits", func(t *testing.T) {
			limit, offset := 5, 0
			result, err := repo.GetByUserID(ctx, userID, limit, offset)
			assert.NoError(t, err)
			assert.Equal(t, 5, len(result))
			for i := range result {
				assert.Equal(t, habits[i].ID, result[i].ID)
				habits[i].CreatedAt = result[i].CreatedAt
				habits[i].UpdatedAt = result[i].UpdatedAt
			}
		})
		t.Run("list limited", func(t *testing.T) {
			limit, offset := 3, 2
			result, err := repo.GetByUserID(ctx, userID, limit, offset)
			assert.NoError(t, err)
			assert.Equal(t, 3, len(result))
			for i := offset; i < 5; i++ {
				assert.Equal(t, *habits[i], *result[i-offset])
			}
		})
		t.Run("list for unknown user", func(t *testing.T) {
			result, err := repo.GetByUserID(ctx, uuid.New(), 10, 0)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(result))
		})
	})
	t.Run("get habit by id", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			h, err := repo.GetByID(ctx, habits[0].ID)
			assert.NoError(t, err)
			assert.Equal(t, *habits[0], *h)
		})
		t.Run("not found", func(t *testing.T) {
			_, err := repo.GetByID(ctx, uuid.New())
			assert.ErrorIs(t, err, errorvalues.ErrHabitNotFound)
		})
	})
	t.Run("update habit", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			h := entity.Habit{
				ID:          habits[0].ID,
				UserID:      userID,
				Title:       "ttt",
				Description: "ddd",
			}
			err := repo.Update(ctx, &h)
			assert.NoError(t, err)
			newHabit, err := repo.GetByID(ctx, h.ID)
			assert.NoError(t, err)
			assert.Equal(t, h.Title, newHabit.Title)
			assert.Equal(t, h.Description, newHabit.Description)
		})
		t.Run("not found", func(t *testing.T) {
			err := repo.Update(ctx, &entity.Habit{
				ID:          uuid.New(),
				Title:       "ttt",
				Description: "ddd",
			})
			assert.ErrorIs(t, err, errorvalues.ErrHabitNotFound)
		})
	})
	t.Run("delete", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			err := repo.Delete(ctx, habits[0].ID)
			assert.NoError(t, err)
			_, err = repo.GetByID(ctx, habits[0].ID)
			assert.ErrorIs(t, err, errorvalues.ErrHabitNotFound)
		})
		t.Run("not found", func(t *testing.T) {
			err := repo.Delete(ctx, uuid.New())
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
	_, err = conn.Exec(`INSERT INTO users (id, name, password_hash) VALUES ($1, $2, $3);`, userID, "test_name", "pass_hash")
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
