package service

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/pkg/entity"
)

type HabitsService struct {
	repo repository.HabitsRepositoryI
}

func NewHabitsService(habitsRepo repository.HabitsRepositoryI) *HabitsService {
	if habitsRepo == nil {
		log.Fatal("provided nil habitsRepo")
	}
	return &HabitsService{
		repo: habitsRepo,
	}
}

func (hs *HabitsService) CreateHabit(ctx context.Context, uid uuid.UUID, req *CreateHabitRequest) (*entity.Habit, error) {
	h := entity.Habit{
		UserID:      uid,
		Title:       req.Title,
		Description: req.Description,
	}
	id, err := hs.repo.Create(ctx, &h)
	if err != nil {
		switch {
		case errors.Is(err, errorvalues.ErrOwnerNotFound):
			return nil, errorvalues.ErrUserNotFound
		case errors.Is(err, errorvalues.ErrUserHasHabit):
			return nil, errorvalues.ErrUserHasHabit
		}
		return nil, errors.New("habits repository error: " + err.Error())
	}
	habit, err := hs.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, errorvalues.ErrHabitNotFound) {
			return nil, err
		}
		return nil, errors.New("habits repository error: " + err.Error())
	}
	return habit, nil
}

func (hs *HabitsService) GetUserHabits(ctx context.Context, uid uuid.UUID, pagination PaginationOpts) ([]*entity.Habit, error) {
	habits, err := hs.repo.GetByUserID(ctx, uid, pagination.Limit, pagination.Offset)
	if err != nil {
		return nil, errors.New("habits repository error: " + err.Error())
	}
	return habits, nil
}

func (hs *HabitsService) DeleteHabit(ctx context.Context, habitID, userID uuid.UUID) error {
	habit, err := hs.repo.GetByID(ctx, habitID)
	if err != nil {
		if errors.Is(err, errorvalues.ErrHabitNotFound) {
			return err
		}
		return errors.New("habits repository error: " + err.Error())
	}
	if habit.UserID != userID {
		return errorvalues.ErrWrongOwner
	}
	err = hs.repo.Delete(ctx, habitID)
	if err != nil {
		if errors.Is(err, errorvalues.ErrHabitNotFound) {
			return err
		}
		return errors.New("habits repository error: " + err.Error())
	}
	return nil
}

func (hs *HabitsService) GetHabit(ctx context.Context, habitID, userID uuid.UUID) (*entity.Habit, error) {
	habit, err := hs.repo.GetByID(ctx, habitID)
	if err != nil {
		if errors.Is(err, errorvalues.ErrHabitNotFound) {
			return nil, err
		}
		return nil, errors.New("habits repository error: " + err.Error())
	}
	if habit.UserID != userID {
		return nil, errorvalues.ErrWrongOwner
	}
	return habit, nil
}
