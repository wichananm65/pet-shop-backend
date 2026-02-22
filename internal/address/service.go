package address

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

func (s *Service) AddAddress(userID int, addressDesc, phone, addressName string) (Address, error) {
    if userID <= 0 {
        return Address{}, ErrNotFound
    }
    return s.repo.AddAddress(userID, addressDesc, phone, addressName)
}

func (s *Service) UpdateAddress(userID int, addressID int, addressDesc, phone, addressName string) (Address, error) {
    if userID <= 0 || addressID <= 0 {
        return Address{}, ErrNotFound
    }
    return s.repo.UpdateAddress(userID, addressID, addressDesc, phone, addressName)
}

func (s *Service) DeleteAddress(userID int, addressID int) error {
    if userID <= 0 || addressID <= 0 {
        return ErrNotFound
    }
    return s.repo.DeleteAddress(userID, addressID)
}
