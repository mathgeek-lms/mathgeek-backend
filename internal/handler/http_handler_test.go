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

func newTestRouter(courseService CourseServiceInterface, lessonService LessonServiceInterface) http.Handler {
	return NewRouter(stubUserService{}, stubTokenService{}, courseService, lessonService)
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

func (stubTokenService) ValidateAccessToken(string) (*service.Claims, error) {
	return nil, errors.New("not implemented")
}

type stubCourseService struct {
	course       model.Course
	courses      []model.Course
	getCourseErr error
	listErr      error
}

func (stubCourseService) CreateCourse(context.Context, model.CreateCourseRequest) (model.Course, error) {
	return model.Course{ID: 1}, nil
}

func (s stubCourseService) GetListCourses(context.Context) ([]model.Course, error) {
	return s.courses, s.listErr
}

func (s stubCourseService) GetCourseByID(context.Context, int64) (model.Course, error) {
	return s.course, s.getCourseErr
}

type stubLessonService struct {
	lesson            model.Lesson
	lessons           []model.Lesson
	getLessonErr      error
	getListLessonsErr error
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
