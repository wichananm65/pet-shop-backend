package favorite

import (
	"errors"
	"sync"

	"github.com/wichananm65/pet-shop-backend/internal/user"
)

var (
	ErrNotFound        = errors.New("user not found")
	ErrAlreadyFavorite = errors.New("product already in favorites")
	ErrNotFavorite     = errors.New("product not in favorites")
)

// Repository provides access to favorite operations.
type Repository interface {
	AddFavorite(userID int, productID int, updatedAt string) ([]int, error)
	RemoveFavorite(userID int, productID int, updatedAt string) ([]int, error)
	GetFavorites(userID int) ([]user.FavoriteProduct, error)
}

// InMemoryRepository is used for tests and local scenarios.
type InMemoryRepository struct {
	mu    sync.RWMutex
	users []user.User
}

func NewInMemoryRepository(seed []user.User) *InMemoryRepository {
	r := &InMemoryRepository{users: make([]user.User, 0, len(seed))}
	for _, u := range seed {
		r.users = append(r.users, u)
	}
	return r
}

func (r *InMemoryRepository) AddFavorite(userID int, productID int, updatedAt string) ([]int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, u := range r.users {
		if u.ID == userID {
			for _, pid := range u.FavoriteProductIDs {
				if pid == productID {
					return nil, ErrAlreadyFavorite
				}
			}
			u.FavoriteProductIDs = append(u.FavoriteProductIDs, productID)
			if updatedAt != "" {
				u.UpdatedAt = updatedAt
			}
			r.users[i] = u
			res := make([]int, len(u.FavoriteProductIDs))
			copy(res, u.FavoriteProductIDs)
			return res, nil
		}
	}
	return nil, ErrNotFound
}

func (r *InMemoryRepository) RemoveFavorite(userID int, productID int, updatedAt string) ([]int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, u := range r.users {
		if u.ID == userID {
			found := false
			newFavs := make([]int, 0, len(u.FavoriteProductIDs))
			for _, pid := range u.FavoriteProductIDs {
				if pid == productID {
					found = true
					continue
				}
				newFavs = append(newFavs, pid)
			}
			if !found {
				return nil, ErrNotFavorite
			}
			u.FavoriteProductIDs = newFavs
			if updatedAt != "" {
				u.UpdatedAt = updatedAt
			}
			r.users[i] = u
			res := make([]int, len(u.FavoriteProductIDs))
			copy(res, u.FavoriteProductIDs)
			return res, nil
		}
	}
	return nil, ErrNotFound
}

func (r *InMemoryRepository) GetFavorites(userID int) ([]user.FavoriteProduct, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.ID == userID {
			out := make([]user.FavoriteProduct, 0, len(u.FavoriteProductIDs))
			for _, pid := range u.FavoriteProductIDs {
				out = append(out, user.FavoriteProduct{ProductID: pid})
			}
			return out, nil
		}
	}
	return nil, ErrNotFound
}
