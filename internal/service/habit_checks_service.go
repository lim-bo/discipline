package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/pkg/entity"
)

type habitChecksService struct {
	habitsRepo repository.HabitsRepository
	checksRepo repository.HabitChecksRepository
}

func NewHabitChecksService(habitsRepo repository.HabitsRepository, checksRepo repository.HabitChecksRepository) HabitChecksService {
	if habitsRepo == nil || checksRepo == nil {
		log.Fatal("on habit checks service provided nil repos")
	}
	return &habitChecksService{
		habitsRepo: habitsRepo,
		checksRepo: checksRepo,
	}
}

func (serv *habitChecksService) CheckHabit(ctx context.Context, habitID, userID uuid.UUID, date time.Time) error {
	habit, err := serv.habitsRepo.GetByID(ctx, habitID)
	if err != nil {
		if errors.Is(err, errorvalues.ErrHabitNotFound) {
			return err
		}
		return fmt.Errorf("repository error: %w", err)
	}
	if habit.UserID != userID {
		return errorvalues.ErrWrongOwner
	}
	if date.After(time.Now()) {
		return errorvalues.ErrCheckDateNotAllowed
	}
	exist, err := serv.checksRepo.Exists(ctx, habitID, date)
	if err != nil {
		return fmt.Errorf("repository error: %w", err)
	}
	if exist {
		return errorvalues.ErrCheckExist
	}
	err = serv.checksRepo.Create(ctx, habitID, date)
	if err != nil {
		return fmt.Errorf("repository error: %w", err)
	}
	return nil
}

func (serv *habitChecksService) UncheckHabit(ctx context.Context, habitID, userID uuid.UUID, date time.Time) error {
	habit, err := serv.habitsRepo.GetByID(ctx, habitID)
	if err != nil {
		if errors.Is(err, errorvalues.ErrHabitNotFound) {
			return err
		}
		return fmt.Errorf("repository error: %w", err)
	}
	if habit.UserID != userID {
		return errorvalues.ErrWrongOwner
	}
	exist, err := serv.checksRepo.Exists(ctx, habitID, date)
	if err != nil {
		return fmt.Errorf("repository error: %w", err)
	}
	if !exist {
		return errorvalues.ErrCheckNotFound
	}
	err = serv.checksRepo.Delete(ctx, habitID, date)
	if err != nil {
		return fmt.Errorf("repository error: %w", err)
	}
	return nil
}

func (serv *habitChecksService) GetHabitChecks(ctx context.Context, habitID, userID uuid.UUID, from, to time.Time) ([]entity.HabitCheck, error) {
	habit, err := serv.habitsRepo.GetByID(ctx, habitID)
	if err != nil {
		if errors.Is(err, errorvalues.ErrHabitNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("repository error: %w", err)
	}
	if habit.UserID != userID {
		return nil, errorvalues.ErrWrongOwner
	}
	checks, err := serv.checksRepo.GetByHabitAndDateRange(ctx, habitID, from, to)
	if err != nil {
		return nil, fmt.Errorf("repository error: %w", err)
	}
	return checks, nil
}

func (serv *habitChecksService) GetHabitStats(ctx context.Context, habitID, userID uuid.UUID) (*entity.HabitStats, error) {
	habit, err := serv.habitsRepo.GetByID(ctx, habitID)
	if err != nil {
		if errors.Is(err, errorvalues.ErrHabitNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("repository error: %w", err)
	}
	if habit.UserID != userID {
		return nil, errorvalues.ErrWrongOwner
	}
	totalChecks, err := serv.checksRepo.CountByHabitID(ctx, habitID)
	if err != nil {
		return nil, fmt.Errorf("repository error: %w", err)
	}
	currentStreak, err := serv.checksRepo.GetCurrentStreak(ctx, habitID)
	if err != nil {
		return nil, fmt.Errorf("repository error: %w", err)
	}
	maxStreak, err := serv.checksRepo.GetMaxStreak(ctx, habitID)
	if err != nil {
		return nil, fmt.Errorf("repository error: %w", err)
	}
	lastCheck, err := serv.checksRepo.GetLastCheckDate(ctx, habitID)
	if err != nil {
		return nil, fmt.Errorf("repository error: %w", err)
	}
	return &entity.HabitStats{
		ID:            habitID,
		TotalChecks:   totalChecks,
		CurrentStreak: currentStreak,
		MaxStreak:     maxStreak,
		LastCheck:     *lastCheck,
	}, nil
}
