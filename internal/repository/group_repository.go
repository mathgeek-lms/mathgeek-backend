package repository

import (
	"context"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
)

type GroupRepository interface {
	GetGroupByID(ctx context.Context, id int64) (model.Group, error)
	GroupExistsByID(ctx context.Context, id int64) (bool, error)
	CreateGroup(ctx context.Context, request model.CreateGroupRequest) (model.Group, error)
}
