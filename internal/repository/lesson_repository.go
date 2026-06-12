package repository

import (
	"context"
	"errors"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
)

type LessonRepository interface {
	CreateLesson(ctx context.Context, request model.CreateLessonRequest) (model.Lesson, error)
	GetListLessonsByCourseID(ctx context.Context, courseID int64) ([]model.Lesson, error)
	GetLessonByID(ctx context.Context, id int64) (model.Lesson, error)
}

var (
	ErrLessonNotFound = errors.New("lesson not found")
	ErrPositionTaken  = errors.New("position taken")
)
