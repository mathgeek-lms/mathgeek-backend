package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	appmiddleware "github.com/mathgeek-lms/mathgeek-backend/internal/middleware"
	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
	"github.com/mathgeek-lms/mathgeek-backend/internal/service"
)

type UserServiceInterface interface {
	CreateUser(ctx context.Context, request model.CreateUserRequest) (model.CreateUserResponse, error)
	LoginUser(ctx context.Context, request model.LoginUserRequest, tokenService service.TokenGenerator) (service.AccessToken, error)
	GetUserByID(ctx context.Context, id int64) (model.CreateUserResponse, error)
}

type TokenServiceInterface interface {
	service.TokenGenerator
	ValidateAccessToken(tokenStr string) (*service.Claims, error)
}

type CourseServiceInterface interface {
	CreateCourse(ctx context.Context, request model.CreateCourseRequest) (model.Course, error)
	GetListCourses(ctx context.Context) ([]model.Course, error)
	GetCourseByID(ctx context.Context, id int64) (model.Course, error)
}

type LessonServiceInterface interface {
	CreateLesson(ctx context.Context, request model.CreateLessonRequest) (model.Lesson, error)
	GetListLessonsByCourseID(ctx context.Context, courseID int64) ([]model.Lesson, error)
	GetLessonByID(ctx context.Context, lessonID int64) (model.Lesson, error)
}

type Handler struct {
	userService   UserServiceInterface
	tokenService  TokenServiceInterface
	courseService CourseServiceInterface
	lessonService LessonServiceInterface
}

func NewRouter(userService UserServiceInterface, tokenService TokenServiceInterface, courseService CourseServiceInterface, lessonService LessonServiceInterface) http.Handler {
	h := &Handler{
		userService:   userService,
		tokenService:  tokenService,
		courseService: courseService,
		lessonService: lessonService,
	}

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.createUserHandler)
			r.Post("/login", h.loginUser)
		})

		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.JWTAuth(h.tokenService))

			r.Get("/me", h.meHandler)
		})
		r.Route("/courses", func(r chi.Router) {
			r.Post("/", h.createCourseHandler)
			r.Get("/", h.getListCoursesHandler)
			r.Get("/{courseID}", h.getCourseByIDHandler)
			r.Get("/{courseID}/lessons", h.getListLessonsByCourseIDHandler)
		})

		r.Route("/lessons", func(r chi.Router) {
			r.Get("/{lessonID}", h.getLessonByIdHandler)
			r.Post("/", h.createLessonHandler)
		})
	})

	return r
}

func (h *Handler) createUserHandler(w http.ResponseWriter, r *http.Request) {
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
			errors.Is(err, service.ErrInvalidPhoneNumber) ||
			errors.Is(err, service.ErrEmptyName) ||
			errors.Is(err, service.ErrEmptyLastName):
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

func (h *Handler) loginUser(w http.ResponseWriter, r *http.Request) {
	var request model.LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "incorrect json: "+err.Error())
		return
	}

	accessToken, err := h.userService.LoginUser(r.Context(), request, h.tokenService)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, accessToken)
}

func (h *Handler) meHandler(w http.ResponseWriter, r *http.Request) {

	claims, ok := appmiddleware.GetClaims(r.Context())
	if !ok || claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userInfo, err := h.userService.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
	}

	writeJSON(w, http.StatusOK, userInfo)
}

func (h *Handler) createCourseHandler(w http.ResponseWriter, r *http.Request) {
	var request model.CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "incorrect json: "+err.Error())
		return
	}

	response, err := h.courseService.CreateCourse(r.Context(), request)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTitle) ||
			errors.Is(err, service.ErrInvalidCourseDuration) ||
			errors.Is(err, repository.ErrTitleTaken) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) getListCoursesHandler(w http.ResponseWriter, r *http.Request) {
	courses, err := h.courseService.GetListCourses(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, courses)
}

func (h *Handler) getCourseByIDHandler(w http.ResponseWriter, r *http.Request) {

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}
	course, err := h.courseService.GetCourseByID(r.Context(), courseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "course not found")
		return
	}
	writeJSON(w, http.StatusOK, course)
}

func (h *Handler) createLessonHandler(w http.ResponseWriter, r *http.Request) {
	var request model.CreateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "incorrect json: "+err.Error())
		return
	}

	lesson, err := h.lessonService.CreateLesson(r.Context(), request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, lesson)
}

func (h *Handler) getListLessonsByCourseIDHandler(w http.ResponseWriter, r *http.Request) {
	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	lessons, err := h.lessonService.GetListLessonsByCourseID(r.Context(), courseID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, lessons)
}

func (h *Handler) getLessonByIdHandler(w http.ResponseWriter, r *http.Request) {
	lessonID, err := strconv.ParseInt(chi.URLParam(r, "lessonID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	lesson, err := h.lessonService.GetLessonByID(r.Context(), lessonID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, lesson)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, model.ErrorResponse{Error: message})
}
