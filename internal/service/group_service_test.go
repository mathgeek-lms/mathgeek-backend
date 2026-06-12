package service

import (
	"context"
	"errors"
	"testing"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/stretchr/testify/require"
)

func TestGroupService_GetGroupByID(t *testing.T) {
	ctx := context.Background()
	expected := model.Group{
		ID:       1,
		CourseID: 2,
		Title:    "Algebra group",
	}
	service := NewGroupService(testGroupRepository{group: expected})

	group, err := service.GetGroupByID(ctx, 1)

	require.NoError(t, err)
	require.Equal(t, expected, group)
}

func TestGroupService_GetGroupByID_Error(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("db error")
	service := NewGroupService(testGroupRepository{getErr: expectedErr})

	_, err := service.GetGroupByID(ctx, 1)

	require.ErrorIs(t, err, expectedErr)
}

func TestGroupService_ExistsGroupByID(t *testing.T) {
	ctx := context.Background()
	service := NewGroupService(testGroupRepository{exists: true})

	exists, err := service.ExistsGroupByID(ctx, 1)

	require.NoError(t, err)
	require.True(t, exists)
}

func TestGroupService_ExistsGroupByID_Error(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("db error")
	service := NewGroupService(testGroupRepository{existsErr: expectedErr})

	exists, err := service.ExistsGroupByID(ctx, 1)

	require.ErrorIs(t, err, expectedErr)
	require.False(t, exists)
}

type testGroupRepository struct {
	group     model.Group
	exists    bool
	getErr    error
	existsErr error
}

func (s testGroupRepository) GetGroupByID(context.Context, int64) (model.Group, error) {
	return s.group, s.getErr
}

func (s testGroupRepository) GroupExistsByID(context.Context, int64) (bool, error) {
	return s.exists, s.existsErr
}
