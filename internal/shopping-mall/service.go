package shoppingmall

// Service provides business logic for shopping-mall endpoints.
type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) List(limit int) []ShoppingMallItem {
	items, err := s.repo.List(limit)
	if err != nil {
		return []ShoppingMallItem{}
	}
	return items
}

func (s *Service) ListLite(limit int) []LiteItem {
	items, err := s.repo.ListLite(limit)
	if err != nil {
		return []LiteItem{}
	}
	return items
}
