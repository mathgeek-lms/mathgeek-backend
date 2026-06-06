package repository

import (
	"context"
	"errors"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, request model.CreateUserRequest) (model.CreateUserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (model.User, error)
	GetUserByID(ctx context.Context, id int64) (model.User, error)
}

var (
	ErrNotFound   = errors.New("not found")
	ErrEmailTaken = errors.New("email already taken")
)
