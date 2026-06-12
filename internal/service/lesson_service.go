package service

import (
	"context"
	"errors"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
)

type LessonService struct {
	repo              repository.LessonRepository
	enrollmentChecker EnrollmentChecker
}

func NewLessonService(repo repository.LessonRepository, enrollmentChecker EnrollmentChecker) *LessonService {
	return &LessonService{repo: repo, enrollmentChecker: enrollmentChecker}
}

func (s *LessonService) CreateLesson(ctx context.Context, request model.CreateLessonRequest) (model.Lesson, error) {
	if request.CourseID == 0 {
		return model.Lesson{}, ErrInvalidCourseId
	}

	if len(request.Title) < 2 || len(request.Title) > 40 {
		return model.Lesson{}, ErrInvalidTitle
	}

	if len(request.Description) < 10 {
		return model.Lesson{}, ErrDescriptionInvalid
	}

	if len(request.Content) < 20 {
		return model.Lesson{}, ErrContentInvalid
	}
	lesson, err := s.repo.CreateLesson(ctx, request)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			return model.Lesson{}, ErrCourseNotFound
		}
		if errors.Is(err, repository.ErrTitleTaken) {
			return model.Lesson{}, ErrTitleTaken
		}
		if errors.Is(err, repository.ErrPositionTaken) {
			return model.Lesson{}, ErrPositionTaken
		}

		return model.Lesson{}, err
	}

	return lesson, err
}

func (s *LessonService) GetListLessonsByCourseID(ctx context.Context, courseID int64) ([]model.Lesson, error) {
	lessons, err := s.repo.GetListLessonsByCourseID(ctx, courseID)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			return nil, ErrCourseNotFound
		}

		return nil, err
	}

	return lessons, nil
}

func (s *LessonService) GetLessonByID(ctx context.Context, lessonID int64) (model.Lesson, error) {
	return s.repo.GetLessonByID(ctx, lessonID)
}

func (s *LessonService) GetLessonForUser(ctx context.Context, userID, lessonID int64, role string) (model.Lesson, error) {
	lesson, err := s.GetLessonByID(ctx, lessonID)
	if err != nil {
		if errors.Is(err, repository.ErrLessonNotFound) {
			return model.Lesson{}, ErrLessonNotFound
		}

		return model.Lesson{}, err
	}

	if role == "ADMIN" {
		return lesson, nil
	} else if role == "STUDENT" {
		isEnrolled, err := s.enrollmentChecker.IsUserEnrolledInCourse(ctx, userID, lesson.CourseID)
		if err != nil {
			return model.Lesson{}, err
		}

		if !isEnrolled {
			return model.Lesson{}, ErrNotEnrolled
		}

		return lesson, nil
	} else {
		return model.Lesson{}, ErrInvalidRole
	}
}

var (
	ErrInvalidCourseId    = errors.New("invalid course id")
	ErrDescriptionInvalid = errors.New("invalid description")
	ErrTitleTaken         = errors.New("title taken")
	ErrPositionTaken      = errors.New("position taken")
	ErrContentInvalid     = errors.New("content invalid")
	ErrInvalidRole        = errors.New("invalid role")
	ErrNotEnrolled        = errors.New("user not enrolled")
	ErrLessonNotFound     = errors.New("lesson not found")
)
