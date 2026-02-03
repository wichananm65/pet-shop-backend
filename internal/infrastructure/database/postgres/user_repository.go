package postgres

import (
	"context"
	"database/sql"
	"errors"

	"pet-shop-backend/internal/domain/entity"
	"pet-shop-backend/internal/domain/repository"
)

// UserRepository is a PostgreSQL implementation of UserRepository.
// This is a skeleton to show infrastructure adapter placement.
type UserRepository struct {
	db *sql.DB
}

var _ repository.UserRepository = (*UserRepository)(nil)

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	return nil, errors.New("postgres repository not implemented")
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	return nil, errors.New("postgres repository not implemented")
}

func (r *UserRepository) List(ctx context.Context) ([]*entity.User, error) {
	return nil, errors.New("postgres repository not implemented")
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	return nil, errors.New("postgres repository not implemented")
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return errors.New("postgres repository not implemented")
}
