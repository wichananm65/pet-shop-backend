package product

import (
	"errors"
	"sync"
)

var (
	ErrNotFound = errors.New("product not found")
)

type Repository interface {
	List() []Product
	GetByID(id int) (Product, error)
	Create(p Product) (Product, error)
	Update(id int, p Product) (Product, error)
	Delete(id int) error
	// Reset replaces all products with the provided list (used for dev / seeding)
	Reset(products []Product) error
}

// InMemoryRepository is a simple in-memory implementation useful for tests and
// seeding local data.
type InMemoryRepository struct {
	mu      sync.RWMutex
	storage []Product
	nextID  int
}

func NewInMemoryRepository(seed []Product) *InMemoryRepository {
	r := &InMemoryRepository{
		storage: make([]Product, 0, len(seed)),
		nextID:  1,
	}

	maxID := 0
	for _, p := range seed {
		r.storage = append(r.storage, p)
		if p.ID > maxID {
			maxID = p.ID
		}
	}

	r.nextID = maxID + 1
	return r
}

func (r *InMemoryRepository) List() []Product {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Product, len(r.storage))
	copy(out, r.storage)
	return out
}

func (r *InMemoryRepository) GetByID(id int) (Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.storage {
		if p.ID == id {
			return p, nil
		}
	}
	return Product{}, ErrNotFound
}

func (r *InMemoryRepository) Create(p Product) (Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p.ID == 0 {
		p.ID = r.nextID
		r.nextID++
	}
	r.storage = append(r.storage, p)
	return p, nil
}

func (r *InMemoryRepository) Update(id int, p Product) (Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.storage {
		if r.storage[i].ID == id {
			p.ID = id
			r.storage[i] = p
			return p, nil
		}
	}
	return Product{}, ErrNotFound
}

func (r *InMemoryRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.storage {
		if r.storage[i].ID == id {
			r.storage = append(r.storage[:i], r.storage[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

// Reset replaces the whole in-memory storage with the provided products.
func (r *InMemoryRepository) Reset(products []Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.storage = make([]Product, 0, len(products))
	maxID := 0
	for _, p := range products {
		if p.ID == 0 {
			p.ID = r.nextID
			r.nextID++
		}
		r.storage = append(r.storage, p)
		if p.ID > maxID {
			maxID = p.ID
		}
	}
	if maxID >= r.nextID {
		r.nextID = maxID + 1
	}
	return nil
}
