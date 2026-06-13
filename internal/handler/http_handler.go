package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/mathgeek-lms/mathgeek-backend/internal/common"
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
	PatchCourseByID(ctx context.Context, id int64, request model.PatchCourseRequest) (model.Course, error)
	IsCourseExistsByID(ctx context.Context, id int64) (bool, error)
}

type LessonServiceInterface interface {
	CreateLesson(ctx context.Context, request model.CreateLessonRequest) (model.Lesson, error)
	GetListLessonsByCourseID(ctx context.Context, courseID int64) ([]model.Lesson, error)
	GetLessonByID(ctx context.Context, lessonID int64) (model.Lesson, error)
	GetLessonForUser(ctx context.Context, userID, lessonID int64, role string) (model.Lesson, error)
	PatchLessonByID(ctx context.Context, lessonID int64, request model.PatchLessonRequest) (model.Lesson, error)
}

type GroupServiceInterface interface {
	GetGroupByID(ctx context.Context, id int64) (model.Group, error)
	ExistsGroupByID(ctx context.Context, id int64) (bool, error)
}

type EnrollmentServiceInterface interface {
	EnrollUserToGroup(ctx context.Context, userID int64, request model.CreateEnrollmentRequest) (model.Enrollment, error)
	ListEnrollmentsByUserID(ctx context.Context, userID int64) ([]model.EnrollmentWithDetails, error)
}

type Handler struct {
	userService       UserServiceInterface
	tokenService      TokenServiceInterface
	courseService     CourseServiceInterface
	lessonService     LessonServiceInterface
	groupService      GroupServiceInterface
	enrollmentService EnrollmentServiceInterface
}

func NewRouter(
	userService UserServiceInterface,
	tokenService TokenServiceInterface,
	courseService CourseServiceInterface,
	lessonService LessonServiceInterface,
	groupService GroupServiceInterface,
	enrollmentService EnrollmentServiceInterface,
) http.Handler {
	h := &Handler{
		userService:       userService,
		tokenService:      tokenService,
		courseService:     courseService,
		lessonService:     lessonService,
		groupService:      groupService,
		enrollmentService: enrollmentService,
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
			r.Get("/me/enrollments", h.getCurrentUserEnrollmentsHandler)
			r.Post("/enrollments", h.enrollmentHandler)
		})
		r.Route("/courses", func(r chi.Router) {
			r.Get("/", h.getListCoursesHandler)
			r.Get("/{courseID}", h.getCourseByIDHandler)
			r.Get("/{courseID}/lessons", h.getListLessonsByCourseIDHandler)
		})

		r.Route("/lessons", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(appmiddleware.JWTAuth(h.tokenService))

				r.Get("/{lessonID}", h.getLessonByIdHandler)
			})
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(appmiddleware.JWTAuth(h.tokenService))
			r.Use(appmiddleware.RequireRole("ADMIN"))

			r.Get("/test", h.adminTestHandler)

			r.Post("/courses", h.createCourseHandler)
			r.Patch("/courses/{courseId}", h.patchCourseHandler)

			r.Post("/lessons", h.createLessonHandler)
			r.Patch("/lessons/{lessonId}", h.patchLessonHandler)
		})
	})

	return r
}

// user & auth handlers

func (h *Handler) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var request model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.WriteError(w, http.StatusBadRequest, "incorrect json: "+err.Error())
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
			common.WriteError(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, service.ErrEmailAlreadyTaken):
			common.WriteError(w, http.StatusConflict, err.Error())
			return
		default:
			common.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	common.WriteJSON(w, http.StatusCreated, user)
}

func (h *Handler) loginUser(w http.ResponseWriter, r *http.Request) {
	var request model.LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.WriteError(w, http.StatusBadRequest, "incorrect json: "+err.Error())
		return
	}

	accessToken, err := h.userService.LoginUser(r.Context(), request, h.tokenService)
	if err != nil {
		common.WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}

	common.WriteJSON(w, http.StatusOK, accessToken)
}

func (h *Handler) meHandler(w http.ResponseWriter, r *http.Request) {

	claims, ok := appmiddleware.GetClaims(r.Context())
	if !ok || claims == nil {
		common.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userInfo, err := h.userService.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			common.WriteError(w, http.StatusUnauthorized, err.Error())
			return
		}

		common.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	common.WriteJSON(w, http.StatusOK, userInfo)
}

// course handlers

func (h *Handler) createCourseHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.GetClaims(r.Context())
	if !ok || claims == nil {
		common.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var request model.CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.WriteError(w, http.StatusBadRequest, "incorrect json: "+err.Error())
		return
	}

	response, err := h.courseService.CreateCourse(r.Context(), request)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTitle) ||
			errors.Is(err, service.ErrInvalidCourseDuration) ||
			errors.Is(err, repository.ErrTitleTaken) {
			common.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		common.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	common.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) getListCoursesHandler(w http.ResponseWriter, r *http.Request) {
	courses, err := h.courseService.GetListCourses(r.Context())
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	common.WriteJSON(w, http.StatusOK, courses)
}

func (h *Handler) getCourseByIDHandler(w http.ResponseWriter, r *http.Request) {

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid course id")
		return
	}
	course, err := h.courseService.GetCourseByID(r.Context(), courseID)
	if err != nil {
		if errors.Is(err, service.ErrCourseNotFound) {
			common.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		common.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	common.WriteJSON(w, http.StatusOK, course)
}

func (h *Handler) patchCourseHandler(w http.ResponseWriter, r *http.Request) {
	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	var request model.PatchCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid json"+err.Error())
		return
	}

	course, err := h.courseService.PatchCourseByID(r.Context(), courseID, request)
	if err != nil {
		if errors.Is(err, service.ErrCourseNotFound) {
			common.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, service.ErrInvalidTitle) ||
			errors.Is(err, service.ErrInvalidCourseDuration) {
			common.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		common.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	common.WriteJSON(w, http.StatusOK, course)
}

// lesson handlers

func (h *Handler) createLessonHandler(w http.ResponseWriter, r *http.Request) {
	var request model.CreateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.WriteError(w, http.StatusBadRequest, "incorrect json: "+err.Error())
		return
	}

	lesson, err := h.lessonService.CreateLesson(r.Context(), request)
	if err != nil {
		if errors.Is(err, service.ErrCourseNotFound) {
			common.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, service.ErrInvalidPosition) ||
			errors.Is(err, service.ErrInvalidTitle) ||
			errors.Is(err, service.ErrInvalidDescription) ||
			errors.Is(err, service.ErrInvalidContent) ||
			errors.Is(err, service.ErrPositionTaken) ||
			errors.Is(err, service.ErrInvalidCourseId) {

			common.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		common.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	common.WriteJSON(w, http.StatusCreated, lesson)
}

func (h *Handler) getListLessonsByCourseIDHandler(w http.ResponseWriter, r *http.Request) {
	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	lessons, err := h.lessonService.GetListLessonsByCourseID(r.Context(), courseID)
	if err != nil {
		if errors.Is(err, service.ErrCourseNotFound) {
			common.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		common.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	common.WriteJSON(w, http.StatusOK, lessonsToListItems(lessons))
}

func (h *Handler) getLessonByIdHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.GetClaims(r.Context())
	if !ok || claims == nil {
		common.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID := claims.UserID

	lessonID, err := strconv.ParseInt(chi.URLParam(r, "lessonID"), 10, 64)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid lesson id")
		return
	}

	lesson, err := h.lessonService.GetLessonForUser(r.Context(), userID, lessonID, claims.Role)
	if err != nil {
		if errors.Is(err, service.ErrLessonNotFound) {
			common.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, service.ErrNotEnrolled) ||
			errors.Is(err, service.ErrInvalidRole) {
			common.WriteError(w, http.StatusForbidden, err.Error())
			return
		}

		common.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.WriteJSON(w, http.StatusOK, lesson)
}

func (h *Handler) patchLessonHandler(w http.ResponseWriter, r *http.Request) {
	lessonID, err := strconv.ParseInt(chi.URLParam(r, "lessonId"), 10, 64)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid lessonId")
		return
	}

	var request model.PatchLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid json"+err.Error())
		return
	}

	_, err = h.lessonService.GetLessonByID(r.Context(), lessonID)
	if err != nil {
		if errors.Is(err, service.ErrLessonNotFound) {
			common.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
	}

	lesson, err := h.lessonService.PatchLessonByID(r.Context(), lessonID, request)
	if err != nil {
		if errors.Is(err, service.ErrCourseNotFound) {
			common.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, service.ErrInvalidPosition) ||
			errors.Is(err, service.ErrInvalidTitle) ||
			errors.Is(err, service.ErrInvalidDescription) ||
			errors.Is(err, service.ErrInvalidContent) ||
			errors.Is(err, service.ErrPositionTaken) ||
			errors.Is(err, service.ErrInvalidCourseId) {

			common.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		common.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	common.WriteJSON(w, http.StatusOK, lesson)
}

// enrollment handlers

func (h *Handler) enrollmentHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.GetClaims(r.Context())
	if !ok || claims == nil {
		common.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var request model.CreateEnrollmentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.WriteError(w, http.StatusBadRequest, "incorrect json: "+err.Error())
		return
	}
	if claims.Role != "STUDENT" {
		common.WriteError(w, http.StatusForbidden, "only student can enroll to group")
		return
	}
	enrollment, err := h.enrollmentService.EnrollUserToGroup(r.Context(), claims.UserID, request)
	if err != nil {
		if errors.Is(err, service.ErrEnrollmentAlreadyExist) {
			common.WriteError(w, http.StatusConflict, err.Error())
			return
		}

		if errors.Is(err, service.ErrGroupNotFound) {
			common.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		common.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	common.WriteJSON(w, http.StatusCreated, enrollment)
}

func (h *Handler) getCurrentUserEnrollmentsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.GetClaims(r.Context())
	if !ok || claims == nil {
		common.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	enrollments, err := h.enrollmentService.ListEnrollmentsByUserID(r.Context(), claims.UserID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	common.WriteJSON(w, http.StatusOK, enrollments)
}

// admin handlers

func (h *Handler) adminTestHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.GetClaims(r.Context())
	if !ok || claims == nil {
		common.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	okResponse := struct {
		Message string
	}{
		Message: "OK",
	}

	common.WriteJSON(w, http.StatusOK, okResponse)
}

func lessonsToListItems(lessons []model.Lesson) []model.LessonListItem {
	response := make([]model.LessonListItem, 0, len(lessons))
	for _, lesson := range lessons {
		response = append(response, model.LessonListItem{
			ID:          lesson.ID,
			CourseID:    lesson.CourseID,
			Title:       lesson.Title,
			Description: lesson.Description,
			Position:    lesson.Position,
			CreatedAt:   lesson.CreatedAt,
			UpdatedAt:   lesson.UpdatedAt,
		})
	}

	return response
}
