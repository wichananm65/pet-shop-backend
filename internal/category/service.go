package category

// Service provides business logic for categories.
type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

// List returns up to `limit` category items.
func (s *Service) List(limit int) []CategoryItem {
	items, err := s.repo.List(limit)
	if err != nil {
		return []CategoryItem{}
	}
	return items
}
