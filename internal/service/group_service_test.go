package service

import (
	"context"
	"errors"
	"testing"
	"time"

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
	service := NewGroupService(testGroupRepository{group: expected}, nil)

	group, err := service.GetGroupByID(ctx, 1)

	require.NoError(t, err)
	require.Equal(t, expected, group)
}

func TestGroupService_GetGroupByID_Error(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("db error")
	service := NewGroupService(testGroupRepository{getErr: expectedErr}, nil)

	_, err := service.GetGroupByID(ctx, 1)

	require.ErrorIs(t, err, expectedErr)
}

func TestGroupService_ExistsGroupByID(t *testing.T) {
	ctx := context.Background()
	service := NewGroupService(testGroupRepository{exists: true}, nil)

	exists, err := service.ExistsGroupByID(ctx, 1)

	require.NoError(t, err)
	require.True(t, exists)
}

func TestGroupService_ExistsGroupByID_Error(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("db error")
	service := NewGroupService(testGroupRepository{existsErr: expectedErr}, nil)

	exists, err := service.ExistsGroupByID(ctx, 1)

	require.ErrorIs(t, err, expectedErr)
	require.False(t, exists)
}

func TestGroupService_CreateGroup(t *testing.T) {
	ctx := context.Background()
	request := model.CreateGroupRequest{
		CourseID: 7,
		Title:    "Algebra group",
	}
	expected := model.Group{
		ID:       1,
		CourseID: request.CourseID,
		Title:    request.Title,
	}
	service := NewGroupService(
		testGroupRepository{createdGroup: expected},
		testCourseChecker{exists: true},
	)

	group, err := service.CreateGroup(ctx, request)

	require.NoError(t, err)
	require.Equal(t, expected, group)
}

func TestGroupService_CreateGroup_CourseNotFound(t *testing.T) {
	ctx := context.Background()
	service := NewGroupService(
		testGroupRepository{},
		testCourseChecker{exists: false},
	)

	_, err := service.CreateGroup(ctx, model.CreateGroupRequest{CourseID: 999, Title: "Algebra group"})

	require.ErrorIs(t, err, ErrCourseNotFound)
}

func TestGroupService_CreateGroup_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)
	beforeFuture := future.Add(-time.Minute)

	tests := []struct {
		name        string
		request     model.CreateGroupRequest
		expectedErr error
	}{
		{
			name: "empty title",
			request: model.CreateGroupRequest{
				CourseID: 7,
				Title:    "",
			},
			expectedErr: ErrInvalidTitle,
		},
		{
			name: "past start date",
			request: model.CreateGroupRequest{
				CourseID:  7,
				Title:     "Algebra group",
				StartDate: &past,
			},
			expectedErr: ErrInvalidStartDate,
		},
		{
			name: "past end date",
			request: model.CreateGroupRequest{
				CourseID: 7,
				Title:    "Algebra group",
				EndDate:  &past,
			},
			expectedErr: ErrInvalidEndDate,
		},
		{
			name: "end date before start date",
			request: model.CreateGroupRequest{
				CourseID:  7,
				Title:     "Algebra group",
				StartDate: &future,
				EndDate:   &beforeFuture,
			},
			expectedErr: ErrInvalidEndDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGroupService(
				testGroupRepository{},
				testCourseChecker{exists: true},
			)

			_, err := service.CreateGroup(ctx, tt.request)

			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

type testGroupRepository struct {
	group        model.Group
	createdGroup model.Group
	exists       bool
	getErr       error
	existsErr    error
	createErr    error
}

func (s testGroupRepository) GetGroupByID(context.Context, int64) (model.Group, error) {
	return s.group, s.getErr
}

func (s testGroupRepository) GroupExistsByID(context.Context, int64) (bool, error) {
	return s.exists, s.existsErr
}

func (s testGroupRepository) CreateGroup(context.Context, model.CreateGroupRequest) (model.Group, error) {
	return s.createdGroup, s.createErr
}

type testCourseChecker struct {
	exists bool
	err    error
}

func (s testCourseChecker) IsCourseExistsByID(context.Context, int64) (bool, error) {
	return s.exists, s.err
}
