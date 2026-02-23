package cart

import (
	"errors"
	"sync"

	"github.com/wichananm65/pet-shop-backend/internal/user"
)

var (
	ErrNotFound = errors.New("user not found")
)

// CartItem describes a product along with its quantity in the cart.
// It reuses FavoriteProduct fields for the product details.
type CartItem struct {
	user.FavoriteProduct
	Quantity int `json:"quantity"`
}

// Repository provides access to cart operations.
// quantities are stored so duplicates are allowed and incremented.
type Repository interface {
	AddToCart(userID int, productID int, qty int, updatedAt string) ([]CartItem, error)
	GetCart(userID int) ([]CartItem, error)
	ClearCart(userID int, updatedAt string) error
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

func (r *InMemoryRepository) AddToCart(userID int, productID int, qty int, updatedAt string) ([]CartItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, u := range r.users {
		if u.ID == userID {
			if u.Cart == nil {
				u.Cart = make(map[int]int)
			}
			u.Cart[productID] += qty
			// remove entry if quantity drops to zero or below
			if u.Cart[productID] <= 0 {
				delete(u.Cart, productID)
			}
			if updatedAt != "" {
				u.UpdatedAt = updatedAt
			}
			r.users[i] = u
			// build response slice
			items := make([]CartItem, 0, len(u.Cart))
			for pid, q := range u.Cart {
				items = append(items, CartItem{FavoriteProduct: user.FavoriteProduct{ProductID: pid}, Quantity: q})
			}
			return items, nil
		}
	}
	return nil, ErrNotFound
}

func (r *InMemoryRepository) GetCart(userID int) ([]CartItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.ID == userID {
			out := make([]CartItem, 0, len(u.Cart))
			for pid, q := range u.Cart {
				out = append(out, CartItem{FavoriteProduct: user.FavoriteProduct{ProductID: pid}, Quantity: q})
			}
			return out, nil
		}
	}
	return nil, ErrNotFound
}

// ClearCart empties a user's cart.
func (r *InMemoryRepository) ClearCart(userID int, updatedAt string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, u := range r.users {
		if u.ID == userID {
			u.Cart = make(map[int]int)
			if updatedAt != "" {
				u.UpdatedAt = updatedAt
			}
			r.users[i] = u
			return nil
		}
	}
	return ErrNotFound
}
