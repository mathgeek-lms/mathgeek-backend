package service

import (
	"context"
	"errors"
	"time"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
)

type GroupService struct {
	repo          repository.GroupRepository
	courseChecker CourseChecker
}

func NewGroupService(repo repository.GroupRepository, courseChecker CourseChecker) *GroupService {
	return &GroupService{repo: repo, courseChecker: courseChecker}
}

func (s *GroupService) GetGroupByID(ctx context.Context, id int64) (model.Group, error) {
	return s.repo.GetGroupByID(ctx, id)
}

func (s *GroupService) ExistsGroupByID(ctx context.Context, id int64) (bool, error) {
	isExist, err := s.repo.GroupExistsByID(ctx, id)
	if err != nil {
		return false, err
	}

	return isExist, nil
}

func (s *GroupService) CreateGroup(ctx context.Context, request model.CreateGroupRequest) (model.Group, error) {
	isCourseExists, err := s.courseChecker.IsCourseExistsByID(ctx, request.CourseID)
	if err != nil {
		return model.Group{}, err
	}

	if !isCourseExists {
		return model.Group{}, ErrCourseNotFound
	}

	if len(request.Title) <= 0 {
		return model.Group{}, ErrInvalidTitle
	}

	if request.StartDate != nil {
		if request.StartDate.Before(time.Now()) {
			return model.Group{}, ErrInvalidStartDate
		}
	}

	if request.EndDate != nil {
		if request.EndDate.Before(time.Now()) {
			return model.Group{}, ErrInvalidEndDate
		}

		if request.StartDate != nil && request.EndDate.Before(*request.StartDate) {
			return model.Group{}, ErrInvalidEndDate
		}
	}

	group, err := s.repo.CreateGroup(ctx, request)
	if err != nil {
		return model.Group{}, err
	}
	return group, nil
}

var (
	ErrInvalidStartDate = errors.New("invalid start_date")
	ErrInvalidEndDate   = errors.New("invalid end_date")
)
