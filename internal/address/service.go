package address

import (
	"errors"
	"time"
)

// Service orchestrates address retrieval.

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAddresses(userID int) ([]Address, error) {
	if userID <= 0 {
		return nil, ErrNotFound
	}
	return s.repo.GetAddresses(userID)
}

func (s *Service) AddAddress(userID int, desc, phone, name string) (Address, error) {
	if userID <= 0 {
		return Address{}, ErrNotFound
	}
	// simple validation
	if desc == "" && name == "" {
		return Address{}, errors.New("addressDesc or addressName required")
	}
	return s.repo.AddAddress(userID, desc, phone, name, time.Now().UTC().Format(time.RFC3339))
}

func (s *Service) UpdateAddress(userID, addressID int, desc, phone, name string) (Address, error) {
	if userID <= 0 || addressID <= 0 {
		return Address{}, ErrNotFound
	}
	if desc == "" && name == "" {
		return Address{}, errors.New("addressDesc or addressName required")
	}
	return s.repo.UpdateAddress(userID, addressID, desc, phone, name, time.Now().UTC().Format(time.RFC3339))
}

func (s *Service) DeleteAddress(userID, addressID int) error {
	if userID <= 0 || addressID <= 0 {
		return ErrNotFound
	}
	return s.repo.DeleteAddress(userID, addressID)
}
