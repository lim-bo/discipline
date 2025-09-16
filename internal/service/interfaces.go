package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/limbo/discipline/pkg/entity"
)

type RegisterRequest struct {
	Name     string `validate:"required,alphanum_underscore,min=3,max=100"`
	Password string `validate:"required,min=8,max=72"`
}

type UserServiceI interface {
	// Validates user's credentials, creates new row in database. Returns user's data with ID
	Register(ctx context.Context, req *RegisterRequest) (*entity.User, error)
	// Compares given credentials. If ok, give back user's data with ID.
	Login(ctx context.Context, name, password string) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByName(ctx context.Context, name string) (*entity.User, error)
	DeleteAccount(ctx context.Context, id uuid.UUID, password string) error
}
