package service

import "context"

type EnrollmentChecker interface {
	IsUserEnrolledInCourse(ctx context.Context, userID, courseID int64) (bool, error)
}
