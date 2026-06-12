package service

import (
	"context"
	"errors"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
)

type EnrollmentService struct {
	repo         repository.EnrollmentRepository
	groupService GroupService
}

func NewEnrollmentService(repo repository.EnrollmentRepository, groupService GroupService) *EnrollmentService {
	return &EnrollmentService{repo: repo, groupService: groupService}
}

func (s *EnrollmentService) EnrollUserToGroup(ctx context.Context, userID int64, request model.CreateEnrollmentRequest) (model.Enrollment, error) {
	isGroupExist, err := s.groupService.ExistsGroupByID(ctx, request.GroupID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Enrollment{}, ErrGroupNotFound
		}
		return model.Enrollment{}, err
	}

	if !isGroupExist {
		return model.Enrollment{}, ErrGroupNotFound
	}

	enrollment, err := s.repo.CreateEnrollment(ctx, userID, request.GroupID)
	if err != nil {
		if errors.Is(err, repository.ErrEnrollmentAlreadyExists) {
			return model.Enrollment{}, ErrEnrollmentAlreadyExist
		}
		return model.Enrollment{}, err
	}

	return enrollment, err
}

func (s *EnrollmentService) ListEnrollmentsByUserID(ctx context.Context, userID int64) ([]model.EnrollmentWithDetails, error) {
	return s.repo.ListEnrollmentsByUserID(ctx, userID)
}

func (s *EnrollmentService) IsUserEnrolledInCourse(ctx context.Context, userID, courseID int64) (bool, error) {
	return s.repo.IsUserEnrolledInCourse(ctx, userID, courseID)
}

var (
	ErrGroupNotFound          = errors.New("group not found")
	ErrEnrollmentAlreadyExist = errors.New("enrollment already exist")
)
