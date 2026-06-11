package service

import (
	"context"
	"errors"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
)

type CourseService struct {
	repo repository.CourseRepository
}

func NewCourseService(repo repository.CourseRepository) *CourseService {
	return &CourseService{repo: repo}
}

func (s *CourseService) CreateCourse(ctx context.Context, request model.CreateCourseRequest) (model.Course, error) {
	if len(request.Title) < 2 || request.Title == "" || len(request.Title) > 40 {
		return model.Course{}, ErrInvalidCourseTitle
	}

	if request.DurationMonths <= 0 {
		return model.Course{}, ErrInvalidCourseDuration
	}

	course, err := s.repo.CreateCourse(ctx, request)
	if err != nil {
		return model.Course{}, err
	}

	return course, err
}

func (s *CourseService) GetListCourses(ctx context.Context) ([]model.Course, error) {
	return s.repo.GetListCourses(ctx)
}

func (s *CourseService) GetCourseByID(ctx context.Context, id int64) (model.Course, error) {
	course, err := s.repo.GetCourseByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Course{}, ErrCourseNotFound
		}

		return model.Course{}, err
	}

	return course, nil
}

var (
	ErrCourseNotFound        = errors.New("course not found")
	ErrInvalidCourseTitle    = errors.New("invalid course title")
	ErrInvalidCourseDuration = errors.New("invalid course duration months")
)
