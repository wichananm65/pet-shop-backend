package cart

// Service orchestrates cart operations.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) AddToCart(userID int, productID int, qty int) ([]CartItem, error) {
	if userID <= 0 || productID <= 0 {
		return nil, ErrNotFound
	}
	// zero qty does nothing, but we still call repo to get current cart
	if qty == 0 {
		return s.repo.GetCart(userID)
	}
	return s.repo.AddToCart(userID, productID, qty, "")
}

func (s *Service) GetCart(userID int) ([]CartItem, error) {
	if userID <= 0 {
		return nil, ErrNotFound
	}
	return s.repo.GetCart(userID)
}

// ClearCart empties a user's cart and returns an error if something goes wrong.
func (s *Service) ClearCart(userID int) error {
	if userID <= 0 {
		return ErrNotFound
	}
	return s.repo.ClearCart(userID, "")
}
