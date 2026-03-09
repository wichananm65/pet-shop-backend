package order

import (
	"errors"
)

// Service provides business logic for orders.
type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) Create(ord Order, userID int) (Order, error) {
	if userID <= 0 {
		return Order{}, errors.New("invalid user")
	}
	if len(ord.Cart) == 0 {
		return Order{}, errors.New("empty cart")
	}
	return s.repo.Create(ord, userID)
}

// ListByIDs retrieves the orders corresponding to the given ids.
func (s *Service) ListByIDs(ids []int) ([]Order, error) {
	if ids == nil {
		return []Order{}, nil
	}
	return s.repo.ListByIDs(ids)
}

// ListByUserID retrieves all orders for a given user directly from the orders table.
func (s *Service) ListByUserID(userID int) ([]Order, error) {
	return s.repo.ListByUserID(userID)
}
