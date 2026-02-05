package user

import (
	"errors"
	"sync"
)

var (
	ErrNotFound           = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email already exists")
)

type Repository interface {
	List() []User
	GetByID(id int) (User, error)
	GetByEmail(email string) (User, error)
	Create(user User) (User, error)
	Update(id int, user User) (User, error)
	Delete(id int) error
}

type InMemoryRepository struct {
	mu     sync.RWMutex
	users  []User
	nextID int
}

func NewInMemoryRepository(seed []User) *InMemoryRepository {
	repo := &InMemoryRepository{
		users:  make([]User, 0, len(seed)),
		nextID: 1,
	}

	maxID := 0
	for _, user := range seed {
		repo.users = append(repo.users, user)
		if user.ID > maxID {
			maxID = user.ID
		}
	}

	repo.nextID = maxID + 1
	return repo
}

func (r *InMemoryRepository) List() []User {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]User, len(r.users))
	copy(users, r.users)
	return users
}

func (r *InMemoryRepository) GetByID(id int) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (r *InMemoryRepository) GetByEmail(email string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (r *InMemoryRepository) Create(user User) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.ID == 0 {
		user.ID = r.nextID
		r.nextID++
	}

	r.users = append(r.users, user)
	return user, nil
}

func (r *InMemoryRepository) Update(id int, userUpdate User) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, user := range r.users {
		if user.ID == id {
			user.Email = userUpdate.Email
			user.FirstName = userUpdate.FirstName
			user.LastName = userUpdate.LastName
			user.Phone = userUpdate.Phone
			user.Gender = userUpdate.Gender
			if userUpdate.Password != "" {
				user.Password = userUpdate.Password
			}
			if userUpdate.CreatedAt != "" {
				user.CreatedAt = userUpdate.CreatedAt
			}
			if userUpdate.UpdatedAt != "" {
				user.UpdatedAt = userUpdate.UpdatedAt
			}
			r.users[i] = user
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (r *InMemoryRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, user := range r.users {
		if user.ID == id {
			r.users = append(r.users[:i], r.users[i+1:]...)
			return nil
		}
	}

	return ErrNotFound
}
