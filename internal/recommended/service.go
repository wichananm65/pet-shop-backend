package recommended

// Service provides business logic for recommended items.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns up to `limit` recommended items ordered by score desc, starting at `offset`.
func (s *Service) List(limit int, offset int) []RecommendedItem {
	items, err := s.repo.List(limit, offset)
	if err != nil {
		return []RecommendedItem{}
	}
	return items
}
