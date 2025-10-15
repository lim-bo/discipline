package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Name         string
	PasswordHash string
}

type Habit struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"uid"`
	Title       string    `json:"title"`
	Description string    `json:"desc"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type HabitCheck struct {
	ID        int
	HabitID   uuid.UUID
	CheckDate time.Time
	CreatedAt time.Time
}

type HabitStats struct {
	ID            uuid.UUID `json:"habit_id"`
	TotalChecks   int       `json:"total_checks"`
	CurrentStreak int       `json:"current_streak"`
	MaxStreak     int       `json:"max_streak"`
	LastCheck     time.Time `json:"last_check,omitempty"`
}
