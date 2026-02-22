package address

import "errors"

var (
    ErrNotFound = errors.New("user not found")
)

type Repository interface {
    GetAddresses(userID int) ([]Address, error)
    AddAddress(userID int, addressDesc, phone, addressName string) (Address, error)
    UpdateAddress(userID int, addressID int, addressDesc, phone, addressName string) (Address, error)
    DeleteAddress(userID int, addressID int) error
}

// InMemoryRepository for tests
type InMemoryRepository struct {
    data map[int][]Address // keyed by userID
}

func NewInMemoryRepository(seed map[int][]Address) *InMemoryRepository {
    return &InMemoryRepository{data: seed}
}

func (r *InMemoryRepository) GetAddresses(userID int) ([]Address, error) {
    if addrs, ok := r.data[userID]; ok {
        return addrs, nil
    }
    return nil, ErrNotFound
}

func (r *InMemoryRepository) AddAddress(userID int, addressDesc, phone, addressName string) (Address, error) {
    if userID <= 0 {
        return Address{}, ErrNotFound
    }
    addr := Address{
        AddressID:   len(r.data[userID]) + 1,
        UserID:      userID,
        AddressDesc: addressDesc,
        Phone:       phone,
        AddressName: addressName,
    }
    r.data[userID] = append(r.data[userID], addr)
    return addr, nil
}

func (r *InMemoryRepository) UpdateAddress(userID int, addressID int, addressDesc, phone, addressName string) (Address, error) {
    if addrs, ok := r.data[userID]; ok {
        for i, a := range addrs {
            if a.AddressID == addressID {
                a.AddressDesc = addressDesc
                a.Phone = phone
                a.AddressName = addressName
                r.data[userID][i] = a
                return a, nil
            }
        }
        return Address{}, ErrNotFound
    }
    return Address{}, ErrNotFound
}

func (r *InMemoryRepository) DeleteAddress(userID int, addressID int) error {
    if addrs, ok := r.data[userID]; ok {
        for i, a := range addrs {
            if a.AddressID == addressID {
                r.data[userID] = append(addrs[:i], addrs[i+1:]...)
                return nil
            }
        }
    }
    return ErrNotFound
}
