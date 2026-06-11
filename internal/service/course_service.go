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
	ErrCourseNotFound = errors.New("course not found")
)
