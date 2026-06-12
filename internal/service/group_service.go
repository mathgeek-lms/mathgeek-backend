package service

import (
	"context"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
)

type GroupService struct {
	repo repository.GroupRepository
}

func NewGroupService(repo repository.GroupRepository) *GroupService {
	return &GroupService{repo: repo}
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
