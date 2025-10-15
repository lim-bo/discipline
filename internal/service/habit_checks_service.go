package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/pkg/entity"
)

type HabitChecksService struct {
	habitsRepo repository.HabitsRepositoryI
	checksRepo repository.HabitChecksRepositoryI
}

func NewHabitChecksService(habitsRepo repository.HabitsRepositoryI, checksRepo repository.HabitChecksRepositoryI) *HabitChecksService {
	if habitsRepo == nil || checksRepo == nil {
		log.Fatal("on habit checks service provided nil repos")
	}
	return &HabitChecksService{
		habitsRepo: habitsRepo,
		checksRepo: checksRepo,
	}
}

func (serv *HabitChecksService) CheckHabit(ctx context.Context, habitID, userID uuid.UUID, date time.Time) error {
	habit, err := serv.habitsRepo.GetByID(ctx, habitID)
	if err != nil {
		if errors.Is(err, errorvalues.ErrHabitNotFound) {
			return err
		}
		return errors.New("repository error: " + err.Error())
	}
	if habit.UserID != userID {
		return errorvalues.ErrWrongOwner
	}
	if date.After(time.Now()) {
		return errorvalues.ErrCheckDateNotAllowed
	}
	exist, err := serv.checksRepo.Exists(ctx, habitID, date)
	if err != nil {
		return errors.New("repository error: " + err.Error())
	}
	if exist {
		return errorvalues.ErrCheckExist
	}
	err = serv.checksRepo.Create(ctx, habitID, date)
	if err != nil {
		return errors.New("repository error: " + err.Error())
	}
	return nil
}

func (serv *HabitChecksService) UncheckHabit(ctx context.Context, habitID, userID uuid.UUID, date time.Time) error {
	habit, err := serv.habitsRepo.GetByID(ctx, habitID)
	if err != nil {
		if errors.Is(err, errorvalues.ErrHabitNotFound) {
			return err
		}
		return errors.New("repository error: " + err.Error())
	}
	if habit.UserID != userID {
		return errorvalues.ErrWrongOwner
	}
	exist, err := serv.checksRepo.Exists(ctx, habitID, date)
	if err != nil {
		return errors.New("repository error: " + err.Error())
	}
	if !exist {
		return errorvalues.ErrCheckNotFound
	}
	err = serv.checksRepo.Delete(ctx, habitID, date)
	if err != nil {
		return errors.New("repository error: " + err.Error())
	}
	return nil
}

func (serv *HabitChecksService) GetHabitChecks(ctx context.Context, habitID, userID uuid.UUID, from, to time.Time) ([]entity.HabitCheck, error) {
	habit, err := serv.habitsRepo.GetByID(ctx, habitID)
	if err != nil {
		if errors.Is(err, errorvalues.ErrHabitNotFound) {
			return nil, err
		}
		return nil, errors.New("repository error: " + err.Error())
	}
	if habit.UserID != userID {
		return nil, errorvalues.ErrWrongOwner
	}
	checks, err := serv.checksRepo.GetByHabitAndDateRange(ctx, habitID, from, to)
	if err != nil {
		return nil, errors.New("repository error: " + err.Error())
	}
	return checks, nil
}

func (serv *HabitChecksService) GetHabitStats(ctx context.Context, habitID, userID uuid.UUID) (*entity.HabitStats, error) {
	habit, err := serv.habitsRepo.GetByID(ctx, habitID)
	if err != nil {
		if errors.Is(err, errorvalues.ErrHabitNotFound) {
			return nil, err
		}
		return nil, errors.New("repository error: " + err.Error())
	}
	if habit.UserID != userID {
		return nil, errorvalues.ErrWrongOwner
	}

	// TO-DO: get back after making streak counting
	return nil, nil
}
