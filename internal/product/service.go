package product

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List() []Product {
	return s.repo.List()
}

func (s *Service) GetByID(id int) (Product, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(p Product) (Product, error) {
	return s.repo.Create(p)
}

func (s *Service) Update(id int, p Product) (Product, error) {
	return s.repo.Update(id, p)
}

func (s *Service) Delete(id int) error {
	return s.repo.Delete(id)
}

// ResetProducts replaces all products with the given list (used for dev / seeding).
func (s *Service) ResetProducts(products []Product) error {
	return s.repo.Reset(products)
}
