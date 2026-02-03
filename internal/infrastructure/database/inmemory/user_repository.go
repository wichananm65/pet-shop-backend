package inmemory

import (
	"context"
	"errors"
	"sync"

	"pet-shop-backend/internal/domain/entity"
	"pet-shop-backend/internal/domain/repository"
)

// UserRepository is an in-memory implementation of UserRepository.
type UserRepository struct {
	mu     sync.RWMutex
	nextID int64
	store  map[int64]*entity.User
}

var _ repository.UserRepository = (*UserRepository)(nil)

func NewUserRepository() *UserRepository {
	return &UserRepository{
		nextID: 1,
		store:  make(map[int64]*entity.User),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	userCopy := *user
	userCopy.ID = r.nextID
	r.nextID++
	r.store[userCopy.ID] = &userCopy

	result := userCopy
	return &result, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.store[id]
	if !ok {
		return nil, errors.New("user not found")
	}

	copy := *user
	return &copy, nil
}

func (r *UserRepository) List(ctx context.Context) ([]*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*entity.User, 0, len(r.store))
	for _, user := range r.store {
		copy := *user
		result = append(result, &copy)
	}
	return result, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[user.ID]; !ok {
		return nil, errors.New("user not found")
	}

	copy := *user
	r.store[user.ID] = &copy
	result := copy
	return &result, nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[id]; !ok {
		return errors.New("user not found")
	}
	delete(r.store, id)
	return nil
}
