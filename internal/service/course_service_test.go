package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCourseService_CreateCourse(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewCourseRepository(t)
	courseService := NewCourseService(repo)

	request := validCreateCourseRequest()
	expectedCourse := model.Course{
		ID:             1,
		Title:          request.Title,
		Description:    *request.Description,
		DurationMonths: request.DurationMonths,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	repo.On("CreateCourse", ctx, request).Return(expectedCourse, nil)

	course, err := courseService.CreateCourse(ctx, request)

	require.NoError(t, err)
	require.Equal(t, expectedCourse, course)
}

func TestCourseService_CreateCourse_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*model.CreateCourseRequest)
		expectedErr error
	}{
		{
			name: "empty title",
			mutate: func(request *model.CreateCourseRequest) {
				request.Title = ""
			},
			expectedErr: ErrInvalidTitle,
		},
		{
			name: "title too short",
			mutate: func(request *model.CreateCourseRequest) {
				request.Title = "A"
			},
			expectedErr: ErrInvalidTitle,
		},
		{
			name: "title too long",
			mutate: func(request *model.CreateCourseRequest) {
				request.Title = "This course title is definitely longer than forty symbols"
			},
			expectedErr: ErrInvalidTitle,
		},
		{
			name: "zero duration",
			mutate: func(request *model.CreateCourseRequest) {
				request.DurationMonths = 0
			},
			expectedErr: ErrInvalidCourseDuration,
		},
		{
			name: "negative duration",
			mutate: func(request *model.CreateCourseRequest) {
				request.DurationMonths = -1
			},
			expectedErr: ErrInvalidCourseDuration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := mocks.NewCourseRepository(t)
			courseService := NewCourseService(repo)
			request := validCreateCourseRequest()
			tt.mutate(&request)

			_, err := courseService.CreateCourse(ctx, request)

			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestCourseService_CreateCourse_RepositoryErrors(t *testing.T) {
	unknownErr := errors.New("database is not happy")
	tests := []struct {
		name        string
		repoErr     error
		expectedErr error
	}{
		{
			name:        "title taken",
			repoErr:     repository.ErrTitleTaken,
			expectedErr: repository.ErrTitleTaken,
		},
		{
			name:        "unknown error",
			repoErr:     unknownErr,
			expectedErr: unknownErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := mocks.NewCourseRepository(t)
			courseService := NewCourseService(repo)
			request := validCreateCourseRequest()

			repo.On("CreateCourse", ctx, request).Return(model.Course{}, tt.repoErr)

			_, err := courseService.CreateCourse(ctx, request)

			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestCourseService_GetListCourses(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewCourseRepository(t)
	courseService := NewCourseService(repo)

	expectedCourses := []model.Course{
		{
			ID:             1,
			Title:          "Algebra",
			Description:    "Learn algebra from scratch.",
			DurationMonths: 3,
		},
		{
			ID:             2,
			Title:          "Geometry",
			Description:    "Learn geometry from scratch.",
			DurationMonths: 4,
		},
	}

	repo.On("GetListCourses", ctx).Return(expectedCourses, nil)

	courses, err := courseService.GetListCourses(ctx)

	require.NoError(t, err)
	require.Equal(t, expectedCourses, courses)
}

func TestCourseService_GetCourseByID(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewCourseRepository(t)
	courseService := NewCourseService(repo)

	expectedCourse := model.Course{
		ID:             1,
		Title:          "Algebra",
		Description:    "Learn algebra from scratch.",
		DurationMonths: 3,
	}

	repo.On("GetCourseByID", ctx, int64(1)).Return(expectedCourse, nil)

	course, err := courseService.GetCourseByID(ctx, 1)

	require.NoError(t, err)
	require.Equal(t, expectedCourse, course)
}

func TestCourseService_GetCourseByID_RepositoryErrors(t *testing.T) {
	unknownErr := errors.New("database is not happy")
	tests := []struct {
		name        string
		repoErr     error
		expectedErr error
	}{
		{
			name:        "not found",
			repoErr:     repository.ErrNotFound,
			expectedErr: ErrCourseNotFound,
		},
		{
			name:        "unknown error",
			repoErr:     unknownErr,
			expectedErr: unknownErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := mocks.NewCourseRepository(t)
			courseService := NewCourseService(repo)

			repo.On("GetCourseByID", ctx, int64(999)).Return(model.Course{}, tt.repoErr)

			_, err := courseService.GetCourseByID(ctx, 999)

			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestCourseService_PatchCourseByID(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewCourseRepository(t)
	courseService := NewCourseService(repo)

	oldCourse := model.Course{
		ID:             7,
		Title:          "Algebra",
		Description:    "Learn algebra from scratch.",
		DurationMonths: 3,
	}
	expectedCourse := model.Course{
		ID:             7,
		Title:          "Advanced Algebra",
		Description:    "Updated course",
		DurationMonths: 4,
	}
	title := "Advanced Algebra"
	description := "Updated course"
	durationMonths := 4
	request := model.PatchCourseRequest{
		Title:          &title,
		Description:    &description,
		DurationMonths: &durationMonths,
	}

	repo.On("GetCourseByID", ctx, int64(7)).Return(oldCourse, nil)
	repo.On("UpdateCourse", ctx, mock.MatchedBy(func(course model.Course) bool {
		return course.ID == 7 &&
			course.Title == title &&
			course.Description == description &&
			course.DurationMonths == durationMonths &&
			!course.UpdatedAt.IsZero()
	})).Return(expectedCourse, nil)

	course, err := courseService.PatchCourseByID(ctx, 7, request)

	require.NoError(t, err)
	require.Equal(t, expectedCourse, course)
}

func TestCourseService_PatchCourseByID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewCourseRepository(t)
	courseService := NewCourseService(repo)
	title := "Advanced Algebra"

	repo.On("GetCourseByID", ctx, int64(999)).Return(model.Course{}, repository.ErrNotFound)

	_, err := courseService.PatchCourseByID(ctx, 999, model.PatchCourseRequest{Title: &title})

	require.ErrorIs(t, err, ErrCourseNotFound)
}

func TestCourseService_PatchCourseByID_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		request     model.PatchCourseRequest
		expectedErr error
	}{
		{
			name:        "empty title",
			request:     model.PatchCourseRequest{Title: stringPtr("")},
			expectedErr: ErrInvalidTitle,
		},
		{
			name:        "bad duration",
			request:     model.PatchCourseRequest{DurationMonths: intPtr(0)},
			expectedErr: ErrInvalidCourseDuration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := mocks.NewCourseRepository(t)
			courseService := NewCourseService(repo)
			oldCourse := model.Course{
				ID:             7,
				Title:          "Algebra",
				Description:    "Learn algebra from scratch.",
				DurationMonths: 3,
			}

			repo.On("GetCourseByID", ctx, int64(7)).Return(oldCourse, nil)

			_, err := courseService.PatchCourseByID(ctx, 7, tt.request)

			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func validCreateCourseRequest() model.CreateCourseRequest {
	description := "Learn algebra from scratch."

	return model.CreateCourseRequest{
		Title:          "Algebra",
		Description:    &description,
		DurationMonths: 3,
	}
}

func stringPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}
