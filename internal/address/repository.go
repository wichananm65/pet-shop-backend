package address

import "errors"

var (
	ErrNotFound = errors.New("user not found")
)

type Repository interface {
	GetAddresses(userID int) ([]Address, error)
	AddAddress(userID int, desc, phone, name string, updatedAt string) (Address, error)
	UpdateAddress(userID, addressID int, desc, phone, name string, updatedAt string) (Address, error)
	DeleteAddress(userID, addressID int) error
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

func (r *InMemoryRepository) AddAddress(userID int, desc, phone, name string, updatedAt string) (Address, error) {
	addrs := r.data[userID]
	newID := 1
	for _, a := range addrs {
		if a.AddressID >= newID {
			newID = a.AddressID + 1
		}
	}
	addr := Address{
		AddressID:   newID,
		UserID:      userID,
		AddressDesc: desc,
		Phone:       phone,
		AddressName: name,
		CreatedAt:   updatedAt,
		UpdatedAt:   updatedAt,
	}
	r.data[userID] = append(addrs, addr)
	return addr, nil
}

func (r *InMemoryRepository) UpdateAddress(userID, addressID int, desc, phone, name string, updatedAt string) (Address, error) {
	addrs := r.data[userID]
	for i, a := range addrs {
		if a.AddressID == addressID {
			a.AddressDesc = desc
			a.Phone = phone
			a.AddressName = name
			a.UpdatedAt = updatedAt
			addrs[i] = a
			r.data[userID] = addrs
			return a, nil
		}
	}
	return Address{}, ErrNotFound
}

func (r *InMemoryRepository) DeleteAddress(userID, addressID int) error {
	addrs := r.data[userID]
	for i, a := range addrs {
		if a.AddressID == addressID {
			r.data[userID] = append(addrs[:i], addrs[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}
