package product

// ServiceInterface defines the subset of functionality used by external
// packages.  It exists primarily to make testing easier and avoid
// depending directly on the concrete Service type.
type ServiceInterface interface {
	List() []Product
	GetByID(id int) (Product, error)
	GetV1ByID(id int) (ProductV1, error)
	ListV1ByIDs(ids []int) ([]ProductV1, error)
	Create(p Product) (Product, error)
	Update(id int, p Product) (Product, error)
	Delete(id int) error
	ListByCategoryID(catID int) []Product
	ResetProducts(products []Product) error
}

var _ ServiceInterface = (*Service)(nil)

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

// GetV1ByID returns the `products`-style product detail for API v1.
func (s *Service) GetV1ByID(id int) (ProductV1, error) {
	return s.repo.GetV1ByID(id)
}

func (s *Service) ListV1ByIDs(ids []int) ([]ProductV1, error) {
	return s.repo.ListV1ByIDs(ids)
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

// ListByCategoryID returns all products associated with the numeric category identifier.
func (s *Service) ListByCategoryID(catID int) []Product {
	return s.repo.ListByCategoryID(catID)
}

// ResetProducts replaces all products with the given list (used for dev / seeding).
func (s *Service) ResetProducts(products []Product) error {
	return s.repo.Reset(products)
}
