package user

import "golang.org/x/crypto/bcrypt"

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
		hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return User{}, err
		}
		user.Password = string(hashed)
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

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	user.Password = string(hashed)
	return s.repo.Create(user)
}

func (s *Service) Authenticate(email, password string) (User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return User{}, ErrInvalidCredentials
	}

	return user, nil
}

func looksLikeBcrypt(value string) bool {
	return len(value) > 4 && value[0:2] == "$2"
}
