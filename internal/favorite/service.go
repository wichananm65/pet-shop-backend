package favorite

import "github.com/wichananm65/pet-shop-backend/internal/user"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) AddFavorite(userID int, productID int) ([]int, error) {
	if userID <= 0 || productID <= 0 {
		return nil, ErrNotFound
	}
	return s.repo.AddFavorite(userID, productID, "")
}

func (s *Service) RemoveFavorite(userID int, productID int) ([]int, error) {
	if userID <= 0 || productID <= 0 {
		return nil, ErrNotFound
	}
	return s.repo.RemoveFavorite(userID, productID, "")
}

func (s *Service) GetFavorites(userID int) ([]user.FavoriteProduct, error) {
	if userID <= 0 {
		return nil, ErrNotFound
	}
	return s.repo.GetFavorites(userID)
}
