package repository

import (
	"context"
	"errors"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
)

type EnrollmentRepository interface {
	CreateEnrollment(ctx context.Context, userID, groupID int64) (model.Enrollment, error)
	ListEnrollmentsByUserID(ctx context.Context, userID int64) ([]model.EnrollmentWithDetails, error)
	IsUserEnrolledInCourse(ctx context.Context, userID, courseID int64) (bool, error)
}

var (
	ErrEnrollmentAlreadyExists = errors.New("enrollment already exists")
)
