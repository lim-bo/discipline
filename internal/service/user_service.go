package service

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/pkg/entity"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo repository.UsersRepositoryI
}

func NewUserService(usersRepo repository.UsersRepositoryI) *UserService {
	return &UserService{
		repo: usersRepo,
	}
}

func (us *UserService) Register(ctx context.Context, req *RegisterRequest) (*entity.User, error) {
	err := validate.Struct(*req)
	if err != nil {
		if validationError, ok := err.(validator.ValidationErrors); ok {
			err = errors.New("validation error: ")
			for _, fieldErr := range validationError {
				err = errors.Join(err, fieldErr)
			}
			return nil, err
		}
		return nil, errors.New("validation unexpected error: " + err.Error())
	}
	passwordHash, err := Hash(req.Password)
	if err != nil {
		return nil, errors.New("hashing password error: " + err.Error())
	}
	err = us.repo.Create(ctx, &entity.User{
		Name:         req.Name,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if errors.Is(err, errorvalues.ErrUserExists) {
			return nil, errors.New("user with such name already exists")
		}
		return nil, errors.New("repository creating error: " + err.Error())
	}
	user, err := us.repo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, errors.New("repository searching error: " + err.Error())
	}
	return user, nil
}

func (us *UserService) Login(ctx context.Context, name, password string) (*entity.User, error) {
	user, err := us.repo.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, errorvalues.ErrUserNotFound) {
			return nil, errors.New("user with given name not found")
		}
		return nil, errors.New("repository searching error: " + err.Error())
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("login failed: wrong password")
	}
	return user, nil
}

func (us *UserService) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := us.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, errorvalues.ErrUserNotFound) {
			return nil, errors.New("user with given id not found")
		}
		return nil, errors.New("repository searching error: " + err.Error())
	}
	return user, nil
}

func (us *UserService) GetByName(ctx context.Context, name string) (*entity.User, error) {
	user, err := us.repo.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, errorvalues.ErrUserNotFound) {
			return nil, errors.New("user with given name not found")
		}
		return nil, errors.New("repository searching error: " + err.Error())
	}
	return user, nil
}

func (us *UserService) DeleteAccount(ctx context.Context, id uuid.UUID, password string) error {
	user, err := us.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, errorvalues.ErrUserNotFound) {
			return errors.New("user with given id not found")
		}
		return errors.New("repository searching error: " + err.Error())
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return errors.New("deletion failed: wrong password")
	}
	err = us.repo.Delete(ctx, user.ID)
	if err != nil {
		if errors.Is(err, errorvalues.ErrUserNotFound) {
			return errors.New("user with given id not found")
		}
		return errors.New("repository deletion error: " + err.Error())
	}
	return nil
}
