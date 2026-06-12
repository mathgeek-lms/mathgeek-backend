package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestEnrollmentRepository_CreateEnrollment(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewEnrollmentRepository(db)
	userID := createTestUser(t, ctx, db, "student@example.com")
	courseID := createTestCourse(t, ctx, db, "Algebra")
	groupID := createTestGroup(t, ctx, db, courseID, "Algebra group")

	enrollment, err := repo.CreateEnrollment(ctx, userID, groupID)

	require.NoError(t, err)
	require.NotZero(t, enrollment.ID)
	require.Equal(t, userID, enrollment.UserID)
	require.Equal(t, groupID, enrollment.GroupID)
	require.Equal(t, "ACTIVE", enrollment.Status)
	require.NotZero(t, enrollment.CreatedAt)
	require.NotZero(t, enrollment.UpdatedAt)
}

func TestEnrollmentRepository_ListEnrollmentsByUserID(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewEnrollmentRepository(db)
	userID := createTestUser(t, ctx, db, "student@example.com")
	courseID := createTestCourse(t, ctx, db, "Algebra")
	groupID := createTestGroup(t, ctx, db, courseID, "Algebra group")
	_, err := repo.CreateEnrollment(ctx, userID, groupID)
	require.NoError(t, err)

	enrollments, err := repo.ListEnrollmentsByUserID(ctx, userID)

	require.NoError(t, err)
	require.Len(t, enrollments, 1)
	require.Equal(t, "ACTIVE", enrollments[0].Status)
	require.Equal(t, groupID, enrollments[0].GroupID)
	require.Equal(t, "Algebra group", enrollments[0].GroupTitle)
	require.Equal(t, courseID, enrollments[0].CourseID)
	require.Equal(t, "Algebra", enrollments[0].CourseTitle)
}

func TestEnrollmentRepository_IsUserEnrolledInCourse(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewEnrollmentRepository(db)
	userID := createTestUser(t, ctx, db, "student@example.com")
	courseID := createTestCourse(t, ctx, db, "Algebra")
	groupID := createTestGroup(t, ctx, db, courseID, "Algebra group")
	_, err := repo.CreateEnrollment(ctx, userID, groupID)
	require.NoError(t, err)

	isEnrolled, err := repo.IsUserEnrolledInCourse(ctx, userID, courseID)

	require.NoError(t, err)
	require.True(t, isEnrolled)
}

func TestEnrollmentRepository_IsUserEnrolledInCourse_NotEnrolled(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewEnrollmentRepository(db)
	userID := createTestUser(t, ctx, db, "student@example.com")
	courseID := createTestCourse(t, ctx, db, "Algebra")

	isEnrolled, err := repo.IsUserEnrolledInCourse(ctx, userID, courseID)

	require.NoError(t, err)
	require.False(t, isEnrolled)
}

func createTestUser(t *testing.T, ctx context.Context, db *pgxpool.Pool, email string) int64 {
	t.Helper()

	var id int64
	err := db.QueryRow(ctx, `
		INSERT INTO users (name, last_name, email, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, "Test", "User", email, "password-hash", "STUDENT").Scan(&id)
	require.NoError(t, err)

	return id
}
