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
	// basic validation already performed by handler, but double-check
	if userID <= 0 {
		return Order{}, errors.New("invalid user")
	}
	if len(ord.Cart) == 0 {
		return Order{}, errors.New("empty cart")
	}
	// TODO: could verify that totals match cart contents
	createdOrder, err := s.repo.Create(ord)
	if err != nil {
		return Order{}, err
	}
	// After order is created, give orderID to user
	// This requires access to user service, which should be handled in handler layer
	return createdOrder, nil
}
