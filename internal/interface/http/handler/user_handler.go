package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"pet-shop-backend/internal/interface/presenter"
	"pet-shop-backend/internal/usecase"
)

// UserHandler adapts HTTP requests to use case calls.
type UserHandler struct {
	usecase   usecase.UserUsecase
	presenter *presenter.UserPresenter
}

func NewUserHandler(usecase usecase.UserUsecase, presenter *presenter.UserPresenter) *UserHandler {
	return &UserHandler{usecase: usecase, presenter: presenter}
}

func (h *UserHandler) ListOrCreate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		users, err := h.usecase.List(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, h.presenter.ToList(users))
	case http.MethodPost:
		var input usecase.CreateUserInput
		if err := decodeJSON(r, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		user, err := h.usecase.Create(r.Context(), input)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, h.presenter.ToResponse(user))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	}
}

func (h *UserHandler) GetUpdateDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid user id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		user, err := h.usecase.GetByID(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, h.presenter.ToResponse(user))
	case http.MethodPatch:
		var input usecase.UpdateUserInput
		if err := decodeJSON(r, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		user, err := h.usecase.Update(r.Context(), id, input)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, h.presenter.ToResponse(user))
	case http.MethodDelete:
		if err := h.usecase.Delete(r.Context(), id); err != nil {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, messageResponse{Message: "deleted"})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	}
}

type messageResponse struct {
	Message string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func decodeJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return errors.New("invalid json body")
	}
	return nil
}

func parseID(path string) (int64, error) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) < 2 {
		return 0, errors.New("missing id")
	}
	return strconv.ParseInt(segments[len(segments)-1], 10, 64)
}
