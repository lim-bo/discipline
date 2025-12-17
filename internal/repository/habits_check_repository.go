package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/pkg/cleanup"
	"github.com/limbo/discipline/pkg/entity"
)

type habitChecksRepository struct {
	conn PgConnection
}

func NewHabitChecksRepo(cfg DBConfig) HabitChecksRepository {
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
	return &habitChecksRepository{
		conn: pool,
	}
}

func NewHabitChecksRepoWithConn(conn PgConnection) HabitChecksRepository {
	err := conn.Ping(context.Background())
	if err != nil {
		log.Fatal("error while pingin connection for habitsRepo: " + err.Error())
	}
	return &habitChecksRepository{
		conn: conn,
	}
}

func (checksRepo *habitChecksRepository) Create(ctx context.Context, habitID uuid.UUID, date time.Time) error {
	_, err := checksRepo.conn.Exec(
		ctx,
		`INSERT INTO habit_checks (habit_id, check_date) VALUES ($1, $2);`,
		habitID,
		date,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			// Unique violation
			case "23505":
				return errorvalues.ErrCheckExist
			// FK violation
			case "23503":
				return errorvalues.ErrHabitNotFound
			}
		}
		return fmt.Errorf("creating check error: %w", err)
	}
	return nil
}

func (checksRepo *habitChecksRepository) Delete(ctx context.Context, habitID uuid.UUID, date time.Time) error {
	ct, err := checksRepo.conn.Exec(
		ctx,
		`DELETE FROM habit_checks WHERE habit_id = $1 AND check_date = $2;`,
		habitID,
		date,
	)
	if err != nil {
		return fmt.Errorf("deleting check error: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return errorvalues.ErrCheckNotFound
	}
	return nil
}

func (checksRepo *habitChecksRepository) Exists(ctx context.Context, habitID uuid.UUID, date time.Time) (bool, error) {
	var exists bool
	row := checksRepo.conn.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM habit_checks WHERE habit_id = $1 AND check_date = $2);`,
		habitID,
		date,
	)
	err := row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("inspecting if check exists error: %w", err)
	}
	return exists, nil
}

func (checksRepo *habitChecksRepository) GetByHabitAndDateRange(ctx context.Context, habitID uuid.UUID, from, to time.Time) ([]entity.HabitCheck, error) {
	rows, err := checksRepo.conn.Query(
		ctx,
		`SELECT id, habit_id, check_date, created_at FROM habit_checks WHERE habit_id = $1 AND check_date >= $2 AND check_date <= $3;`,
		habitID,
		from,
		to,
	)
	if err != nil {
		return nil, fmt.Errorf("getting checks for period error: %w", err)
	}
	result := make([]entity.HabitCheck, 0, 2)
	for rows.Next() {
		check := entity.HabitCheck{}
		err = rows.Scan(&check.ID, &check.HabitID, &check.CheckDate, &check.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("check row parsing error: %w", err)
		}
		result = append(result, check)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("unexpected check rows error: %w", err)
	}
	return result, nil
}

func (checksRepo *habitChecksRepository) GetLastCheckDate(ctx context.Context, habitID uuid.UUID) (*time.Time, error) {
	row := checksRepo.conn.QueryRow(
		ctx,
		`SELECT check_date FROM habit_checks WHERE habit_id = $1 ORDER BY check_date DESC LIMIT 1;`,
		habitID,
	)
	var date time.Time
	if err := row.Scan(&date); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting last check date error: %w", err)
	}
	return &date, nil
}

func (checksRepo *habitChecksRepository) CountByHabitID(ctx context.Context, habitID uuid.UUID) (int, error) {
	row := checksRepo.conn.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM habit_checks WHERE habit_id = $1;`,
		habitID,
	)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("error counting checks: %w", err)
	}
	return count, nil
}

func (checksRepo *habitChecksRepository) GetCurrentStreak(ctx context.Context, habitID uuid.UUID) (int, error) {
	query := `
		WITH RECURSIVE streak_calc AS (
			SELECT check_date, 1 as streak_lenght
			FROM habit_checks
			WHERE habit_id = $1 AND check_date = CURRENT_DATE

			UNION ALL

			SELECT h.check_date, s.streak_lenght + 1
			FROM streak_calc s
			JOIN habit_checks h ON h.habit_id = $1 AND h.check_date = s.check_date - INTERVAL '1 day'
		)
		SELECT COALESCE(MAX(streak_lenght), 0) FROM streak_calc;
	`
	var count int
	if err := checksRepo.conn.QueryRow(ctx, query, habitID).Scan(&count); err != nil {
		return 0, fmt.Errorf("error getting current streak: %w", err)
	}
	return count, nil
}

func (checksRepo *habitChecksRepository) GetMaxStreak(ctx context.Context, habitID uuid.UUID) (int, error) {
	query := `
		WITH RECURSIVE all_streaks AS (
			SELECT check_date AS start_date, check_date AS current_date, 1 AS streak_lenght
			FROM habit_checks
			WHERE habit_id = $1

			UNION ALL

			SELECT a.start_date, h.check_date, a.streak_lenght + 1
			FROM all_streaks a
			JOIN habit_checks h ON h.habit_id = $1
				AND h.check_date = a.current_date + INTERVAL '1 day'
		)

		SELECT COALESCE(MAX(streak_lenght), 0)
		FROM all_streaks;
	`
	var count int
	if err := checksRepo.conn.QueryRow(ctx, query, habitID).Scan(&count); err != nil {
		return 0, fmt.Errorf("error getting max streak: %w", err)
	}
	return count, nil
}
