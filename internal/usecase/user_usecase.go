package usecase

import (
	"context"

	"pet-shop-backend/internal/domain/entity"
)

// UserUsecase exposes application-level operations for User.
type UserUsecase interface {
	Create(ctx context.Context, input CreateUserInput) (*entity.User, error)
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	List(ctx context.Context) ([]*entity.User, error)
	Update(ctx context.Context, id int64, input UpdateUserInput) (*entity.User, error)
	Delete(ctx context.Context, id int64) error
}

// CreateUserInput carries data required to create a user.
type CreateUserInput struct {
	Email     string
	Firstname string
	Lastname  string
	Phone     string
	Gender    string
	AvatarPic string
}

// UpdateUserInput carries data required to update a user.
type UpdateUserInput struct {
	Email     string
	Firstname string
	Lastname  string
	Phone     string
	Gender    string
	AvatarPic string
}
