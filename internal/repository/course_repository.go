package repository

import (
	"context"
	"errors"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
)

type CourseRepository interface {
	CreateCourse(ctx context.Context, request model.CreateCourseRequest) (model.Course, error)
	GetListCourses(ctx context.Context) ([]model.Course, error)
	GetCourseByID(ctx context.Context, id int64) (model.Course, error)
}

var (
	ErrCourseNotFound = errors.New("course not found")
)
