package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/service"
	"github.com/stretchr/testify/require"
)

func TestRouter_WriteCourseAndLessonRoutesAreNotPublished(t *testing.T) {
	router := newTestRouter(stubCourseService{}, stubLessonService{})

	tests := []struct {
		name   string
		path   string
		body   string
		method string
	}{
		{
			name:   "create course",
			path:   "/api/v1/courses/",
			body:   `{"title":"Algebra","description":"Learn algebra","duration_months":3}`,
			method: http.MethodPost,
		},
		{
			name:   "create lesson",
			path:   "/api/v1/lessons/",
			body:   `{"course_id":1,"title":"Linear equations","description":"Learn equations","content":"Long enough lesson content","position":1}`,
			method: http.MethodPost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))

			router.ServeHTTP(recorder, request)

			require.Contains(t, []int{http.StatusNotFound, http.StatusMethodNotAllowed}, recorder.Code)
		})
	}
}

func TestGetCourseByIDHandler_NotFoundReturns404(t *testing.T) {
	router := newTestRouter(stubCourseService{
		getCourseErr: service.ErrCourseNotFound,
	}, stubLessonService{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/courses/999", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGetListLessonsByCourseIDHandler_CourseNotFoundReturns404(t *testing.T) {
	router := newTestRouter(stubCourseService{}, stubLessonService{
		getListLessonsErr: service.ErrCourseNotFound,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/courses/999/lessons", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGetListLessonsByCourseIDHandler_ReturnsShortLessonResponse(t *testing.T) {
	router := newTestRouter(stubCourseService{}, stubLessonService{
		lessons: []model.Lesson{
			{
				ID:          1,
				CourseID:    7,
				Title:       "Linear equations",
				Description: "Learn how to solve simple linear equations.",
				Content:     "This full content must stay out of list responses.",
				Position:    1,
			},
		},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/courses/7/lessons", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response []map[string]any
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	require.Len(t, response, 1)
	require.Equal(t, "Linear equations", response[0]["title"])
	require.NotContains(t, response[0], "content")
}

func TestGetLessonByIDHandler_EnrolledUserCanGetLesson(t *testing.T) {
	router := newTestRouter(stubCourseService{}, stubLessonService{
		lesson: model.Lesson{
			ID:       1,
			CourseID: 7,
			Title:    "Linear equations",
			Content:  "Full lesson content for enrolled students.",
		},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/lessons/1", nil)
	request.Header.Set("Authorization", "Bearer valid-token")

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	require.Equal(t, "Linear equations", response["title"])
	require.Equal(t, "Full lesson content for enrolled students.", response["content"])
}

func TestGetLessonByIDHandler_NonEnrolledUserGets403(t *testing.T) {
	router := newTestRouter(stubCourseService{}, stubLessonService{
		lesson: model.Lesson{
			ID:       1,
			CourseID: 7,
			Title:    "Linear equations",
		},
		getLessonForUserErr: service.ErrNotEnrolled,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/lessons/1", nil)
	request.Header.Set("Authorization", "Bearer valid-token")

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestGetLessonByIDHandler_NoTokenGets401(t *testing.T) {
	router := newTestRouter(stubCourseService{}, stubLessonService{
		lesson: model.Lesson{
			ID:       1,
			CourseID: 7,
			Title:    "Linear equations",
		},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/lessons/1", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestGetCurrentUserEnrollmentsHandler_NoTokenReturns401(t *testing.T) {
	router := newTestRouter(stubCourseService{}, stubLessonService{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me/enrollments", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestGetCurrentUserEnrollmentsHandler_ReturnsEnrollmentDetails(t *testing.T) {
	router := newTestRouterWithEnrollment(stubCourseService{}, stubLessonService{}, stubEnrollmentService{
		enrollments: []model.EnrollmentWithDetails{
			{
				ID:          9,
				Status:      "ACTIVE",
				GroupID:     3,
				GroupTitle:  "Algebra Basics Group A",
				CourseID:    2,
				CourseTitle: "Algebra Basics",
			},
		},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me/enrollments", nil)
	request.Header.Set("Authorization", "Bearer valid-token")

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response []map[string]any
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	require.Len(t, response, 1)
	require.Equal(t, "ACTIVE", response[0]["status"])
	require.Equal(t, "Algebra Basics Group A", response[0]["group_title"])
	require.Equal(t, "Algebra Basics", response[0]["course_title"])
	require.NotContains(t, response[0], "password")
	require.NotContains(t, response[0], "password_hash")
}

func TestEnrollmentHandler_NonStudentRolesGet403(t *testing.T) {
	router := newTestRouter(stubCourseService{}, stubLessonService{})

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "admin",
			token: "admin-token",
		},
		{
			name:  "unknown role",
			token: "unknown-role-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/enrollments", strings.NewReader(`{"group_id":7}`))
			request.Header.Set("Authorization", "Bearer "+tt.token)

			router.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusForbidden, recorder.Code)
		})
	}
}

func TestAdminTestHandler_AccessControl(t *testing.T) {
	router := newTestRouter(stubCourseService{}, stubLessonService{})

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{
			name:       "no token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "student token",
			token:      "valid-token",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "admin token",
			token:      "admin-token",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/test", nil)
			if tt.token != "" {
				request.Header.Set("Authorization", "Bearer "+tt.token)
			}

			router.ServeHTTP(recorder, request)

			require.Equal(t, tt.wantStatus, recorder.Code)
		})
	}
}

func TestAdminCreateCourseHandler_StudentGets403(t *testing.T) {
	router := newTestRouter(stubCourseService{}, stubLessonService{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/courses", strings.NewReader(`{"title":"Algebra","description":"Learn algebra","duration_months":3}`))
	request.Header.Set("Authorization", "Bearer valid-token")

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestAdminCreateCourseHandler_AdminCreatesCourse(t *testing.T) {
	router := newTestRouter(stubCourseService{
		createdCourse: model.Course{
			ID:             7,
			Title:          "Algebra",
			Description:    "Learn algebra",
			DurationMonths: 3,
		},
	}, stubLessonService{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/courses", strings.NewReader(`{"title":"Algebra","description":"Learn algebra","duration_months":3}`))
	request.Header.Set("Authorization", "Bearer admin-token")

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusCreated, recorder.Code)

	var response map[string]any
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	require.Equal(t, float64(7), response["id"])
	require.Equal(t, "Algebra", response["title"])
	require.Equal(t, "Learn algebra", response["description"])
	require.Equal(t, float64(3), response["duration_months"])
}

func TestAdminCreateCourseHandler_ValidationErrorsReturn400(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		createErr error
	}{
		{
			name:      "empty title",
			body:      `{"title":"","description":"Learn algebra","duration_months":3}`,
			createErr: service.ErrInvalidTitle,
		},
		{
			name:      "bad duration",
			body:      `{"title":"Algebra","description":"Learn algebra","duration_months":0}`,
			createErr: service.ErrInvalidCourseDuration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newTestRouter(stubCourseService{
				createErr: tt.createErr,
			}, stubLessonService{})

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/courses", strings.NewReader(tt.body))
			request.Header.Set("Authorization", "Bearer admin-token")

			router.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusBadRequest, recorder.Code)
		})
	}
}

func newTestRouter(courseService CourseServiceInterface, lessonService LessonServiceInterface) http.Handler {
	return newTestRouterWithEnrollment(courseService, lessonService, stubEnrollmentService{})
}

func newTestRouterWithEnrollment(courseService CourseServiceInterface, lessonService LessonServiceInterface, enrollmentService EnrollmentServiceInterface) http.Handler {
	return NewRouter(stubUserService{}, stubTokenService{}, courseService, lessonService, stubGroupService{}, enrollmentService)
}

type stubUserService struct{}

func (stubUserService) CreateUser(context.Context, model.CreateUserRequest) (model.CreateUserResponse, error) {
	return model.CreateUserResponse{}, nil
}

func (stubUserService) LoginUser(context.Context, model.LoginUserRequest, service.TokenGenerator) (service.AccessToken, error) {
	return service.AccessToken{}, nil
}

func (stubUserService) GetUserByID(context.Context, int64) (model.CreateUserResponse, error) {
	return model.CreateUserResponse{}, nil
}

type stubTokenService struct{}

func (stubTokenService) GenerateAccessToken(int64, string) (service.AccessToken, error) {
	return service.AccessToken{}, nil
}

func (stubTokenService) ValidateAccessToken(token string) (*service.Claims, error) {
	switch token {
	case "valid-token":
		return &service.Claims{UserID: 42, Role: "STUDENT"}, nil
	case "admin-token":
		return &service.Claims{UserID: 42, Role: "ADMIN"}, nil
	case "unknown-role-token":
		return &service.Claims{UserID: 42, Role: "TEACHER"}, nil
	}

	return nil, errors.New("not implemented")
}

type stubCourseService struct {
	course        model.Course
	courses       []model.Course
	createdCourse model.Course
	createErr     error
	getCourseErr  error
	listErr       error
}

func (s stubCourseService) CreateCourse(context.Context, model.CreateCourseRequest) (model.Course, error) {
	if s.createErr != nil {
		return model.Course{}, s.createErr
	}
	if s.createdCourse.ID != 0 || s.createdCourse.Title != "" {
		return s.createdCourse, nil
	}
	return model.Course{ID: 1}, nil
}

func (s stubCourseService) GetListCourses(context.Context) ([]model.Course, error) {
	return s.courses, s.listErr
}

func (s stubCourseService) GetCourseByID(context.Context, int64) (model.Course, error) {
	return s.course, s.getCourseErr
}

type stubLessonService struct {
	lesson              model.Lesson
	lessons             []model.Lesson
	getLessonErr        error
	getLessonForUserErr error
	getListLessonsErr   error
}

func (stubLessonService) CreateLesson(context.Context, model.CreateLessonRequest) (model.Lesson, error) {
	return model.Lesson{ID: 1}, nil
}

func (s stubLessonService) GetListLessonsByCourseID(context.Context, int64) ([]model.Lesson, error) {
	return s.lessons, s.getListLessonsErr
}

func (s stubLessonService) GetLessonByID(context.Context, int64) (model.Lesson, error) {
	return s.lesson, s.getLessonErr
}

func (s stubLessonService) GetLessonForUser(context.Context, int64, int64, string) (model.Lesson, error) {
	return s.lesson, s.getLessonForUserErr
}

type stubGroupService struct{}

func (stubGroupService) GetGroupByID(context.Context, int64) (model.Group, error) {
	return model.Group{}, nil
}

func (stubGroupService) ExistsGroupByID(context.Context, int64) (bool, error) {
	return true, nil
}

type stubEnrollmentService struct {
	enrollment  model.Enrollment
	enrollments []model.EnrollmentWithDetails
	isEnrolled  bool
	createErr   error
	listErr     error
	enrolledErr error
}

func (s stubEnrollmentService) EnrollUserToGroup(context.Context, int64, model.CreateEnrollmentRequest) (model.Enrollment, error) {
	return s.enrollment, s.createErr
}

func (s stubEnrollmentService) ListEnrollmentsByUserID(context.Context, int64) ([]model.EnrollmentWithDetails, error) {
	return s.enrollments, s.listErr
}

func (s stubEnrollmentService) IsUserEnrolledInCourse(context.Context, int64, int64) (bool, error) {
	return s.isEnrolled, s.enrolledErr
}
