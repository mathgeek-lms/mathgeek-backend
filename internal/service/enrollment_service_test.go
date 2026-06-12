package service

import (
	"context"
	"testing"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestEnrollmentService_EnrollUserToGroup(t *testing.T) {
	ctx := context.Background()
	expected := model.Enrollment{
		ID:      1,
		UserID:  42,
		GroupID: 7,
		Status:  "ACTIVE",
	}
	groupService := NewGroupService(testGroupRepository{exists: true})
	service := NewEnrollmentService(testEnrollmentRepository{enrollment: expected}, *groupService)

	enrollment, err := service.EnrollUserToGroup(ctx, 42, model.CreateEnrollmentRequest{GroupID: 7})

	require.NoError(t, err)
	require.Equal(t, expected, enrollment)
}

func TestEnrollmentService_EnrollUserToGroup_GroupNotFound(t *testing.T) {
	ctx := context.Background()
	groupService := NewGroupService(testGroupRepository{exists: false})
	service := NewEnrollmentService(testEnrollmentRepository{}, *groupService)

	_, err := service.EnrollUserToGroup(ctx, 42, model.CreateEnrollmentRequest{GroupID: 7})

	require.ErrorIs(t, err, ErrGroupNotFound)
}

func TestEnrollmentService_EnrollUserToGroup_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	groupService := NewGroupService(testGroupRepository{exists: true})
	service := NewEnrollmentService(testEnrollmentRepository{
		createErr: repository.ErrEnrollmentAlreadyExists,
	}, *groupService)

	_, err := service.EnrollUserToGroup(ctx, 42, model.CreateEnrollmentRequest{GroupID: 7})

	require.ErrorIs(t, err, ErrEnrollmentAlreadyExist)
}

func TestEnrollmentService_ListEnrollmentsByUserID(t *testing.T) {
	ctx := context.Background()
	expected := []model.EnrollmentWithDetails{
		{
			ID:          1,
			Status:      "ACTIVE",
			GroupID:     7,
			GroupTitle:  "Algebra group",
			CourseID:    3,
			CourseTitle: "Algebra",
		},
	}
	groupService := NewGroupService(testGroupRepository{})
	service := NewEnrollmentService(testEnrollmentRepository{enrollments: expected}, *groupService)

	enrollments, err := service.ListEnrollmentsByUserID(ctx, 42)

	require.NoError(t, err)
	require.Equal(t, expected, enrollments)
}

func TestEnrollmentService_IsUserEnrolledInCourse(t *testing.T) {
	ctx := context.Background()
	groupService := NewGroupService(testGroupRepository{})
	service := NewEnrollmentService(testEnrollmentRepository{isEnrolled: true}, *groupService)

	isEnrolled, err := service.IsUserEnrolledInCourse(ctx, 42, 3)

	require.NoError(t, err)
	require.True(t, isEnrolled)
}

type testEnrollmentRepository struct {
	enrollment  model.Enrollment
	enrollments []model.EnrollmentWithDetails
	isEnrolled  bool
	createErr   error
	listErr     error
	enrolledErr error
}

func (s testEnrollmentRepository) CreateEnrollment(context.Context, int64, int64) (model.Enrollment, error) {
	return s.enrollment, s.createErr
}

func (s testEnrollmentRepository) ListEnrollmentsByUserID(context.Context, int64) ([]model.EnrollmentWithDetails, error) {
	return s.enrollments, s.listErr
}

func (s testEnrollmentRepository) IsUserEnrolledInCourse(context.Context, int64, int64) (bool, error) {
	return s.isEnrolled, s.enrolledErr
}
