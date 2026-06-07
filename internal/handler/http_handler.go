package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	appmiddleware "github.com/mathgeek-lms/mathgeek-backend/internal/middleware"
	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/service"
)

type UserServiceInterface interface {
	CreateUser(ctx context.Context, request model.CreateUserRequest) (model.CreateUserResponse, error)
	LoginUser(ctx context.Context, request model.LoginUserRequest) (service.AccessToken, error)
}

type TokenServiceInterface interface {
	GenerateAccessToken(userID int64, email, role string) (service.AccessToken, error)
	ValidateAccessToken(tokenStr string) (*service.Claims, error)
}

type UserHandler struct {
	userService  UserServiceInterface
	tokenService TokenServiceInterface
}

func NewRouter(userService UserServiceInterface, tokenService TokenServiceInterface) http.Handler {
	h := &UserHandler{
		userService:  userService,
		tokenService: tokenService,
	}

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.createUser)
			r.Post("/login", h.loginUser)
		})

		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.JWTAuth(h.tokenService))

			r.Get("/me", h.meHandler)
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

func (h *UserHandler) meHandler(w http.ResponseWriter, r *http.Request) {

	claims, ok := appmiddleware.GetClaims(r.Context())
	if !ok || claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, claims)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, model.ErrorResponse{Error: message})
}
