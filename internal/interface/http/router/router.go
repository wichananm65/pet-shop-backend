package router

import (
	"net/http"
	"strings"

	"pet-shop-backend/internal/interface/http/handler"
)

// New builds an HTTP router without framework lock-in.
func New(userHandler *handler.UserHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/v1/users", userHandler.ListOrCreate)
	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v1/users/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		userHandler.GetUpdateDelete(w, r)
	})

	return mux
}
