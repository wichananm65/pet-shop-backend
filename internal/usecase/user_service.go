package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"pet-shop-backend/internal/domain/entity"
	"pet-shop-backend/internal/domain/repository"
)

// UserService implements UserUsecase with repository dependency.
type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*entity.User, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	first := strings.TrimSpace(input.Firstname)
	last := strings.TrimSpace(input.Lastname)

	if email == "" || first == "" || last == "" {
		return nil, errors.New("email, firstname and lastname are required")
	}

	user := &entity.User{
		Email:     email,
		Firstname: first,
		Lastname:  last,
		Phone:     strings.TrimSpace(input.Phone),
		Gender:    strings.TrimSpace(input.Gender),
		AvatarPic: strings.TrimSpace(input.AvatarPic),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	return s.repo.Create(ctx, user)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) List(ctx context.Context) ([]*entity.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) Update(ctx context.Context, id int64, input UpdateUserInput) (*entity.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user id")
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(input.Email) != "" {
		user.Email = strings.ToLower(strings.TrimSpace(input.Email))
	}
	if strings.TrimSpace(input.Firstname) != "" {
		user.Firstname = strings.TrimSpace(input.Firstname)
	}
	if strings.TrimSpace(input.Lastname) != "" {
		user.Lastname = strings.TrimSpace(input.Lastname)
	}
	if strings.TrimSpace(input.Phone) != "" {
		user.Phone = strings.TrimSpace(input.Phone)
	}
	if strings.TrimSpace(input.Gender) != "" {
		user.Gender = strings.TrimSpace(input.Gender)
	}
	if strings.TrimSpace(input.AvatarPic) != "" {
		user.AvatarPic = strings.TrimSpace(input.AvatarPic)
	}

	user.UpdatedAt = time.Now().UTC()
	return s.repo.Update(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid user id")
	}
	return s.repo.Delete(ctx, id)
}
