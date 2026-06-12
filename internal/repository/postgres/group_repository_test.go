package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestGroupRepository_GetGroupByID(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewGroupRepository(db)
	courseID := createTestCourse(t, ctx, db, "Algebra")
	groupID := createTestGroup(t, ctx, db, courseID, "Algebra group")

	group, err := repo.GetGroupByID(ctx, groupID)

	require.NoError(t, err)
	require.Equal(t, groupID, group.ID)
	require.Equal(t, courseID, group.CourseID)
	require.Equal(t, "Algebra group", group.Title)
	require.NotNil(t, group.StartDate)
	require.NotNil(t, group.EndDate)
	require.NotZero(t, group.CreatedAt)
	require.NotZero(t, group.UpdatedAt)
}

func TestGroupRepository_GetGroupByID_NotFound(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewGroupRepository(db)

	_, err := repo.GetGroupByID(ctx, 999)

	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestGroupRepository_GroupExistsByID(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewGroupRepository(db)
	courseID := createTestCourse(t, ctx, db, "Algebra")
	groupID := createTestGroup(t, ctx, db, courseID, "Algebra group")

	exists, err := repo.GroupExistsByID(ctx, groupID)

	require.NoError(t, err)
	require.True(t, exists)
}

func TestGroupRepository_GroupExistsByID_NotFound(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewGroupRepository(db)

	exists, err := repo.GroupExistsByID(ctx, 999)

	require.NoError(t, err)
	require.False(t, exists)
}

func createTestGroup(t *testing.T, ctx context.Context, db *pgxpool.Pool, courseID int64, title string) int64 {
	t.Helper()

	startDate := time.Now()
	endDate := startDate.AddDate(0, 1, 0)

	var id int64
	err := db.QueryRow(ctx, `
		INSERT INTO groups (course_id, title, start_date, end_date)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, courseID, title, startDate, endDate).Scan(&id)
	require.NoError(t, err)

	return id
}
