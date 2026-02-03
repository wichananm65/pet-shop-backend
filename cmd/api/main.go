package main

import (
	"log"
	"net/http"

	"pet-shop-backend/internal/infrastructure/config"
	"pet-shop-backend/internal/infrastructure/database/inmemory"
	httpHandler "pet-shop-backend/internal/interface/http/handler"
	"pet-shop-backend/internal/interface/http/router"
	"pet-shop-backend/internal/interface/presenter"
	"pet-shop-backend/internal/usecase"
)

// main wires dependencies (dependency injection) and starts the HTTP server.
func main() {
	cfg := config.Load()

	userRepo := inmemory.NewUserRepository()
	userPresenter := presenter.NewUserPresenter()
	userUsecase := usecase.NewUserService(userRepo)
	userHandler := httpHandler.NewUserHandler(userUsecase, userPresenter)

	r := router.New(userHandler)

	log.Printf("starting server on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
