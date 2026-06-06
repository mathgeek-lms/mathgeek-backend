package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/service"
)

type UserServiceInterface interface {
	CreateUser(ctx context.Context, request model.CreateUserRequest) (model.CreateUserResponse, error)
	LoginUser(ctx context.Context, request model.LoginUserRequest) (service.AccessToken, error)
}

type UserHandler struct {
	userService UserServiceInterface
}

func NewRouter(userService UserServiceInterface) http.Handler {
	h := &UserHandler{userService: userService}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.createUser)
			r.Post("/login", h.loginUser)
		})
	})

	return r
}

func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var request model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "incorrect json: "+err.Error())
		return
	}

	user, err := h.userService.CreateUser(r.Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail) ||
			errors.Is(err, service.ErrPasswordTooShort) ||
			errors.Is(err, service.ErrInvalidPhoneNumber):
			writeError(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, service.ErrEmailAlreadyTaken):
			writeError(w, http.StatusConflict, err.Error())
			return
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) loginUser(w http.ResponseWriter, r *http.Request) {
	var request model.LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "incorrect json: "+err.Error())
		return
	}

	accessToken, err := h.userService.LoginUser(r.Context(), request)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, accessToken)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, model.ErrorResponse{Error: message})
}
