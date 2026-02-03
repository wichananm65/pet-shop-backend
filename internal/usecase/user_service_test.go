package usecase

import (
	"context"
	"testing"

	"pet-shop-backend/internal/infrastructure/database/inmemory"
)

func TestUserService_CreateAndGet(t *testing.T) {
	repo := inmemory.NewUserRepository()
	svc := NewUserService(repo)

	created, err := svc.Create(context.Background(), CreateUserInput{
		Email:     "user@example.com",
		Firstname: "Jane",
		Lastname:  "Doe",
	})
	if err != nil {
		t.Fatalf("expected create to succeed, got error: %v", err)
	}

	fetched, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected get to succeed, got error: %v", err)
	}

	if fetched.Email != created.Email {
		t.Fatalf("expected email %s, got %s", created.Email, fetched.Email)
	}
}
