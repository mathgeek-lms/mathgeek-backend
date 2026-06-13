package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository/mocks"
	"github.com/stretchr/testify/require"
)

func TestLessonService_CreateLesson(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLessonRepository(t)
	lessonService := NewLessonService(repo, nil, nil)

	request := validCreateLessonRequest()
	expectedLesson := model.Lesson{
		ID:          1,
		CourseID:    request.CourseID,
		Title:       request.Title,
		Description: request.Description,
		Content:     request.Content,
		Position:    int(request.Position),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	repo.On("CreateLesson", ctx, request).Return(expectedLesson, nil)

	lesson, err := lessonService.CreateLesson(ctx, request)

	require.NoError(t, err)
	require.Equal(t, expectedLesson, lesson)
}

func TestLessonService_CreateLesson_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*model.CreateLessonRequest)
		expectedErr error
	}{
		{
			name: "invalid course id",
			mutate: func(request *model.CreateLessonRequest) {
				request.CourseID = 0
			},
			expectedErr: ErrInvalidCourseId,
		},
		{
			name: "title too short",
			mutate: func(request *model.CreateLessonRequest) {
				request.Title = "A"
			},
			expectedErr: ErrInvalidTitle,
		},
		{
			name: "title too long",
			mutate: func(request *model.CreateLessonRequest) {
				request.Title = "This lesson title is definitely longer than forty symbols"
			},
			expectedErr: ErrInvalidTitle,
		},
		{
			name: "description too short",
			mutate: func(request *model.CreateLessonRequest) {
				request.Description = "too short"
			},
			expectedErr: ErrInvalidDescription,
		},
		{
			name: "content too short",
			mutate: func(request *model.CreateLessonRequest) {
				request.Content = "short content"
			},
			expectedErr: ErrInvalidContent,
		},
		{
			name: "invalid position",
			mutate: func(request *model.CreateLessonRequest) {
				request.Position = 0
			},
			expectedErr: ErrInvalidPosition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := mocks.NewLessonRepository(t)
			lessonService := NewLessonService(repo, nil, nil)
			request := validCreateLessonRequest()
			tt.mutate(&request)

			_, err := lessonService.CreateLesson(ctx, request)

			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestLessonService_CreateLesson_RepositoryErrors(t *testing.T) {
	unknownErr := errors.New("database is not happy")
	tests := []struct {
		name        string
		repoErr     error
		expectedErr error
	}{
		{
			name:        "course not found",
			repoErr:     repository.ErrCourseNotFound,
			expectedErr: ErrCourseNotFound,
		},
		{
			name:        "title taken",
			repoErr:     repository.ErrTitleTaken,
			expectedErr: ErrTitleTaken,
		},
		{
			name:        "position taken",
			repoErr:     repository.ErrPositionTaken,
			expectedErr: ErrPositionTaken,
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
			repo := mocks.NewLessonRepository(t)
			lessonService := NewLessonService(repo, nil, nil)
			request := validCreateLessonRequest()

			repo.On("CreateLesson", ctx, request).Return(model.Lesson{}, tt.repoErr)

			_, err := lessonService.CreateLesson(ctx, request)

			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestLessonService_GetListLessonsByCourseID(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLessonRepository(t)
	lessonService := NewLessonService(repo, nil, nil)

	expectedLessons := []model.Lesson{
		{
			ID:       1,
			CourseID: 7,
			Title:    "Lesson 1",
			Position: 1,
		},
		{
			ID:       2,
			CourseID: 7,
			Title:    "Lesson 2",
			Position: 2,
		},
	}

	repo.On("GetListLessonsByCourseID", ctx, int64(7)).Return(expectedLessons, nil)

	lessons, err := lessonService.GetListLessonsByCourseID(ctx, 7)

	require.NoError(t, err)
	require.Equal(t, expectedLessons, lessons)
}

func TestLessonService_GetListLessonsByCourseID_CourseNotFound(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLessonRepository(t)
	lessonService := NewLessonService(repo, nil, nil)

	repo.On("GetListLessonsByCourseID", ctx, int64(999)).Return(nil, repository.ErrCourseNotFound)

	_, err := lessonService.GetListLessonsByCourseID(ctx, 999)

	require.ErrorIs(t, err, ErrCourseNotFound)
}

func TestLessonService_GetLessonByID(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLessonRepository(t)
	lessonService := NewLessonService(repo, nil, nil)

	expectedLesson := model.Lesson{
		ID:       3,
		CourseID: 7,
		Title:    "Lesson 3",
		Position: 3,
	}

	repo.On("GetLessonByID", ctx, int64(3)).Return(expectedLesson, nil)

	lesson, err := lessonService.GetLessonByID(ctx, 3)

	require.NoError(t, err)
	require.Equal(t, expectedLesson, lesson)
}

func TestLessonService_GetLessonForUser_EnrolledStudent(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLessonRepository(t)
	lessonService := NewLessonService(repo, stubEnrollmentChecker{isEnrolled: true}, nil)

	expectedLesson := model.Lesson{
		ID:       3,
		CourseID: 7,
		Title:    "Lesson 3",
		Position: 3,
	}

	repo.On("GetLessonByID", ctx, int64(3)).Return(expectedLesson, nil)

	lesson, err := lessonService.GetLessonForUser(ctx, 42, 3, "STUDENT")

	require.NoError(t, err)
	require.Equal(t, expectedLesson, lesson)
}

func TestLessonService_GetLessonForUser_NonEnrolledStudent(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLessonRepository(t)
	lessonService := NewLessonService(repo, stubEnrollmentChecker{isEnrolled: false}, nil)

	expectedLesson := model.Lesson{
		ID:       3,
		CourseID: 7,
		Title:    "Lesson 3",
		Position: 3,
	}

	repo.On("GetLessonByID", ctx, int64(3)).Return(expectedLesson, nil)

	_, err := lessonService.GetLessonForUser(ctx, 42, 3, "STUDENT")

	require.ErrorIs(t, err, ErrNotEnrolled)
}

func TestLessonService_GetLessonForUser_AdminDoesNotNeedEnrollment(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLessonRepository(t)
	lessonService := NewLessonService(repo, nil, nil)

	expectedLesson := model.Lesson{
		ID:       3,
		CourseID: 7,
		Title:    "Lesson 3",
		Position: 3,
	}

	repo.On("GetLessonByID", ctx, int64(3)).Return(expectedLesson, nil)

	lesson, err := lessonService.GetLessonForUser(ctx, 42, 3, "ADMIN")

	require.NoError(t, err)
	require.Equal(t, expectedLesson, lesson)
}

func TestLessonService_GetLessonForUser_InvalidRole(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLessonRepository(t)
	lessonService := NewLessonService(repo, nil, nil)

	expectedLesson := model.Lesson{
		ID:       3,
		CourseID: 7,
		Title:    "Lesson 3",
		Position: 3,
	}

	repo.On("GetLessonByID", ctx, int64(3)).Return(expectedLesson, nil)

	_, err := lessonService.GetLessonForUser(ctx, 42, 3, "TEACHER")

	require.ErrorIs(t, err, ErrInvalidRole)
}

func validCreateLessonRequest() model.CreateLessonRequest {
	return model.CreateLessonRequest{
		CourseID:    1,
		Title:       "Linear equations",
		Description: "Learn how to solve simple linear equations.",
		Content:     "This lesson explains equations step by step with examples.",
		Position:    1,
	}
}

type stubEnrollmentChecker struct {
	isEnrolled bool
	err        error
}

func (s stubEnrollmentChecker) IsUserEnrolledInCourse(context.Context, int64, int64) (bool, error) {
	return s.isEnrolled, s.err
}
