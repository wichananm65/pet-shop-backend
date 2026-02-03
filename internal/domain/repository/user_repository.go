package repository

import (
	"context"

	"pet-shop-backend/internal/domain/entity"
)

// UserRepository defines persistence behavior for the User entity.
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	List(ctx context.Context) ([]*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	Delete(ctx context.Context, id int64) error
}
