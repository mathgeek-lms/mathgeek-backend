package service

import (
	"context"
	"errors"
	"time"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
)

type LessonService struct {
	repo              repository.LessonRepository
	enrollmentChecker EnrollmentChecker
	courseChecker     CourseChecker
}

func NewLessonService(repo repository.LessonRepository, enrollmentChecker EnrollmentChecker, courseChecker CourseChecker) *LessonService {
	return &LessonService{repo: repo, enrollmentChecker: enrollmentChecker, courseChecker: courseChecker}
}

func (s *LessonService) CreateLesson(ctx context.Context, request model.CreateLessonRequest) (model.Lesson, error) {
	if request.CourseID == 0 {
		return model.Lesson{}, ErrInvalidCourseId
	}

	if len(request.Title) < 2 || len(request.Title) > 40 {
		return model.Lesson{}, ErrInvalidTitle
	}

	if len(request.Description) < 10 {
		return model.Lesson{}, ErrInvalidDescription
	}

	if len(request.Content) < 20 {
		return model.Lesson{}, ErrInvalidContent
	}

	if request.Position <= 0 {
		return model.Lesson{}, ErrInvalidPosition
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

func (s *LessonService) PatchLessonByID(ctx context.Context, lessonID int64, request model.PatchLessonRequest) (model.Lesson, error) {
	oldLesson, err := s.repo.GetLessonByID(ctx, lessonID)
	if err != nil {
		if errors.Is(err, repository.ErrLessonNotFound) {
			return model.Lesson{}, ErrLessonNotFound
		}

		return model.Lesson{}, err
	}

	diffCount := 0

	if request.CourseID != nil {
		isCourseExists, err := s.courseChecker.IsCourseExistsByID(ctx, *request.CourseID)

		if err != nil {
			if errors.Is(err, ErrCourseNotFound) {
				return model.Lesson{}, err
			}

			return model.Lesson{}, err
		}

		if !isCourseExists {
			return model.Lesson{}, ErrCourseNotFound
		}
		oldLesson.CourseID = *request.CourseID
		diffCount++
	}

	if request.Title != nil {
		if len(*request.Title) >= 2 && len(*request.Title) <= 40 {
			oldLesson.Title = *request.Title
			diffCount++
		} else {
			return model.Lesson{}, ErrInvalidTitle
		}
	}

	if request.Description != nil {
		if len(*request.Description) > 10 {
			oldLesson.Description = *request.Description
			diffCount++
		} else {
			return model.Lesson{}, ErrInvalidDescription
		}
	}

	if request.Content != nil {
		if len(*request.Content) > 20 {
			oldLesson.Content = *request.Content
			diffCount++
		} else {
			return model.Lesson{}, ErrInvalidContent
		}
	}

	if request.Position != nil && *request.Position != oldLesson.Position {
		isPositionTaken, err := s.repo.IsLessonPositionTaken(ctx, oldLesson.CourseID, *request.Position)

		if *request.Position <= 0 {
			return model.Lesson{}, ErrInvalidPosition
		}

		if err == nil && !isPositionTaken {
			oldLesson.Position = *request.Position
			diffCount++
		} else {
			return model.Lesson{}, ErrPositionTaken
		}
	}

	if diffCount <= 0 {
		return oldLesson, nil
	}

	oldLesson.UpdatedAt = time.Now()

	response, err := s.repo.UpdateLesson(ctx, oldLesson)
	if err != nil {
		if errors.Is(err, repository.ErrTitleTaken) {
			return model.Lesson{}, ErrTitleTaken
		}
		if errors.Is(err, repository.ErrPositionTaken) {
			return model.Lesson{}, ErrPositionTaken
		}

		return model.Lesson{}, err
	}

	return response, err

}

var (
	ErrInvalidCourseId    = errors.New("invalid course id")
	ErrInvalidDescription = errors.New("invalid description")
	ErrTitleTaken         = errors.New("title taken")
	ErrInvalidPosition    = errors.New("position must be positive")
	ErrPositionTaken      = errors.New("position taken")
	ErrInvalidContent     = errors.New("content invalid")
	ErrInvalidRole        = errors.New("invalid role")
	ErrNotEnrolled        = errors.New("user not enrolled")
	ErrLessonNotFound     = errors.New("lesson not found")
)
