package repository

import (
	"context"
	"errors"
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

type HabitChecksRepository struct {
	conn PgConnection
}

func NewHabitChecksRepo(cfg DBConfig) *HabitChecksRepository {
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
	return &HabitChecksRepository{
		conn: pool,
	}
}

func NewHabitChecksRepoWithConn(conn PgConnection) *HabitChecksRepository {
	err := conn.Ping(context.Background())
	if err != nil {
		log.Fatal("error while pingin connection for habitsRepo: " + err.Error())
	}
	return &HabitChecksRepository{
		conn: conn,
	}
}

func (checksRepo *HabitChecksRepository) Create(ctx context.Context, habitID uuid.UUID, date time.Time) error {
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
		return errors.New("creating check error: " + err.Error())
	}
	return nil
}

func (checksRepo *HabitChecksRepository) Delete(ctx context.Context, habitID uuid.UUID, date time.Time) error {
	ct, err := checksRepo.conn.Exec(
		ctx,
		`DELETE FROM habit_checks WHERE habit_id = $1 AND check_date = $2;`,
		habitID,
		date,
	)
	if err != nil {
		return errors.New("deleting check error: " + err.Error())
	}
	if ct.RowsAffected() == 0 {
		return errorvalues.ErrCheckNotFound
	}
	return nil
}

func (checksRepo *HabitChecksRepository) Exists(ctx context.Context, habitID uuid.UUID, date time.Time) (bool, error) {
	var exists bool
	row := checksRepo.conn.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM habit_checks WHERE habitID = $1 AND check_date = $2);`,
		habitID,
		date,
	)
	err := row.Scan(&exists)
	if err != nil {
		return false, errors.New("inspecting if check exists error: " + err.Error())
	}
	return exists, nil
}

func (checksRepo *HabitChecksRepository) GetByHabitAndDateRange(ctx context.Context, habitID uuid.UUID, from, to time.Time) ([]entity.HabitCheck, error) {
	rows, err := checksRepo.conn.Query(
		ctx,
		`SELECT id, habit_id, check_date, created_at FROM habit_checks WHERE habitID = $1 AND check_date >= $2 AND check_date <= $3;`,
		habitID,
		from,
		to,
	)
	if err != nil {
		return nil, errors.New("getting checks for period error: " + err.Error())
	}
	result := make([]entity.HabitCheck, 0, 2)
	for rows.Next() {
		check := entity.HabitCheck{}
		err = rows.Scan(check.ID, check.HabitID, check.CheckDate, check.CreatedAt)
		if err != nil {
			return nil, errors.New("check row parsing error: " + err.Error())
		}
		result = append(result, check)
	}
	if rows.Err() != nil {
		return nil, errors.New("unexpected check rows error: " + err.Error())
	}
	return result, nil
}

func (checksRepo *HabitChecksRepository) GetLastCheckDate(ctx context.Context, habitID uuid.UUID) (*time.Time, error) {
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
		return nil, errors.New("getting last check date error: " + err.Error())
	}
	return &date, nil
}

func (checksRepo *HabitChecksRepository) CountByHabitID(ctx context.Context, habitID uuid.UUID) (int, error) {
	row := checksRepo.conn.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM habit_checks WHERE habit_id = $1;`,
		habitID,
	)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, errors.New("error counting checks: " + err.Error())
	}
	return count, nil
}
