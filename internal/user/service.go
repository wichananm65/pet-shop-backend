package user

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List() []User {
	return s.repo.List()
}

func (s *Service) GetByID(id int) (User, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(user User) (User, error) {
	if user.Password != "" && !looksLikeBcrypt(user.Password) {
		hashed, err := hashPassword(user.Password)
		if err != nil {
			return User{}, err
		}
		user.Password = hashed
	}

	return s.repo.Create(user)
}

func (s *Service) Update(id int, user User) (User, error) {
	return s.repo.Update(id, user)
}

func (s *Service) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *Service) Register(user User) (User, error) {
	if _, err := s.repo.GetByEmail(user.Email); err == nil {
		return User{}, ErrEmailExists
	} else if err != ErrNotFound {
		return User{}, err
	}

	hashed, err := hashPassword(user.Password)
	if err != nil {
		return User{}, err
	}
	user.Password = hashed
	created, err := s.repo.Create(user)
	if err != nil {
		return User{}, err
	}

	// cart table removed — no cart row needs to be created during registration
	return created, nil
}

func (s *Service) Authenticate(email, password string) (User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}
	// If stored password looks like a bcrypt hash, validate via bcrypt
	if looksLikeBcrypt(user.Password) {
		if !passwordMatchesHash(user.Password, password) {
			return User{}, ErrInvalidCredentials
		}
		return user, nil
	}

	// Legacy plaintext password: compare directly and upgrade to bcrypt on success
	if user.Password == password {
		// attempt to upgrade stored password to bcrypt
		if hashed, err := hashPassword(password); err == nil {
			// update user record with hashed password and updatedAt timestamp
			user.Password = hashed
			user.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			// best-effort update; ignore update error to avoid blocking login
			_, _ = s.repo.Update(user.ID, user)
		}
		return user, nil
	}

	return User{}, ErrInvalidCredentials
}

func looksLikeBcrypt(value string) bool {
	return len(value) > 4 && value[0:2] == "$2"
}

func hashPassword(value string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func passwordMatchesHash(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
