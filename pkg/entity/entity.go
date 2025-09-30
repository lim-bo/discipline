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
