package banner

// Service provides business logic for banners.
type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

// List returns up to `limit` banner items.
func (s *Service) List(limit int) []BannerItem {
	items, err := s.repo.List(limit)
	if err != nil {
		return []BannerItem{}
	}
	return items
}
