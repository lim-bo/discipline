package repository

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/pkg/cleanup"
	"github.com/limbo/discipline/pkg/entity"
)

type HabitsRepository struct {
	conn PgConnection
}

func NewHabitsRepo(cfg DBConfig) *HabitsRepository {
	pool, err := pgxpool.New(context.Background(), cfg.ConnString())
	if err != nil {
		log.Fatal("creating connection for usersRepo error: " + err.Error())
	}
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal("error while pinging connection for usersRepo: " + err.Error())
	}
	cleanup.Register(&cleanup.Job{
		Name: "closing pgxpool",
		F: func() error {
			pool.Close()
			return nil
		},
	})
	return &HabitsRepository{
		conn: pool,
	}
}

func NewHabitsRepoWithConn(conn PgConnection) *HabitsRepository {
	err := conn.Ping(context.Background())
	if err != nil {
		log.Fatal("error while pingin connection for habitsRepo: " + err.Error())
	}
	return &HabitsRepository{
		conn: conn,
	}
}

func (hr *HabitsRepository) Create(ctx context.Context, habit *entity.Habit) error {
	_, err := hr.conn.Exec(ctx, `INSERT INTO habits (user_id, title, description) VALUES ($1, $2)`,
		habit.UserID,
		habit.Title,
		habit.Description,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			// Unique violation
			case "23505":
				return errorvalues.ErrUserHasHabit
			}
		}
		return errors.New("creating habit db error: " + err.Error())
	}
	return nil
}

func (hr *HabitsRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Habit, error) {
	var habit entity.Habit
	habit.ID = id
	row := hr.conn.QueryRow(ctx, `SELECT user_id, title, description, created_at, updated_at FROM habits WHERE id = $1;`, id)
	if err := row.Scan(&habit.UserID, &habit.Title, &habit.Description, &habit.CreatedAt, &habit.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorvalues.ErrHabitNotFound
		}
		return nil, errors.New("getting habit by id error: " + err.Error())
	}
	return &habit, nil

}

func (hr *HabitsRepository) GetByUserID(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*entity.Habit, error) {
	habits := make([]*entity.Habit, 0)
	rows, err := hr.conn.Query(ctx, `SELECT id, user_id, title, description, created_at, updated_at 
		FROM habits WHERE user_id = $1 IMIT $2 OFFSET $3;`, uid, limit, offset)
	if err != nil {
		return nil, errors.New("getting habits by uid error: " + err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		h := entity.Habit{}
		err = rows.Scan(&h.ID, &h.UserID, &h.Title, &h.Description, &h.CreatedAt, &h.UpdatedAt)
		if err != nil {
			return nil, errors.New("unmarhalling habit error: " + err.Error())
		}
		habits = append(habits, &h)
	}
	if rows.Err() != nil {
		return nil, errors.New("unexpected error after scanning: " + err.Error())
	}
	return habits, nil
}

func (hr *HabitsRepository) Update(ctx context.Context, habit *entity.Habit) error {
	ct, err := hr.conn.Exec(ctx, `UPDATE habits SET title = $1, description = $2, updated_at = NOW() WHERE id = $3;`,
		habit.Title, habit.Description, habit.ID,
	)
	if err != nil {
		return errors.New("error updating habit: " + err.Error())
	}
	if ct.RowsAffected() == 0 {
		return errorvalues.ErrHabitNotFound
	}
	return nil
}

func (hr *HabitsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := hr.conn.Exec(ctx, `DELETE FROM habits WHERE id = $1;`, id)
	if err != nil {
		return errors.New("error deleting habit: " + err.Error())
	}
	if ct.RowsAffected() == 0 {
		return errorvalues.ErrHabitNotFound
	}
	return nil
}
