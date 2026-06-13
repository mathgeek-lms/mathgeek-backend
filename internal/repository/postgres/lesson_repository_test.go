package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestLessonRepository_CreateLesson(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewLessonRepository(db)
	courseID := createTestCourse(t, ctx, db, "Algebra")

	request := model.CreateLessonRequest{
		CourseID:    courseID,
		Title:       "Linear equations",
		Description: "Learn how to solve simple linear equations.",
		Content:     "This lesson explains equations step by step with examples.",
		Position:    1,
	}

	lesson, err := repo.CreateLesson(ctx, request)

	require.NoError(t, err)
	require.NotZero(t, lesson.ID)
	require.Equal(t, request.CourseID, lesson.CourseID)
	require.Equal(t, request.Title, lesson.Title)
	require.Equal(t, request.Description, lesson.Description)
	require.Equal(t, request.Content, lesson.Content)
	require.Equal(t, int(request.Position), lesson.Position)
	require.NotZero(t, lesson.CreatedAt)
	require.NotZero(t, lesson.UpdatedAt)

	var saved model.Lesson
	err = db.QueryRow(ctx, `
		SELECT id, course_id, title, COALESCE(description, ''), COALESCE(content, ''), position, created_at, updated_at
		FROM lessons
		WHERE id = $1
	`, lesson.ID).Scan(
		&saved.ID,
		&saved.CourseID,
		&saved.Title,
		&saved.Description,
		&saved.Content,
		&saved.Position,
		&saved.CreatedAt,
		&saved.UpdatedAt,
	)
	require.NoError(t, err)
	require.Equal(t, lesson.ID, saved.ID)
	require.Equal(t, request.CourseID, saved.CourseID)
	require.Equal(t, request.Title, saved.Title)
	require.Equal(t, request.Description, saved.Description)
	require.Equal(t, request.Content, saved.Content)
	require.Equal(t, int(request.Position), saved.Position)
	require.WithinDuration(t, lesson.CreatedAt, saved.CreatedAt, time.Second)
	require.WithinDuration(t, lesson.UpdatedAt, saved.UpdatedAt, time.Second)
}

func TestLessonRepository_CreateLesson_CourseNotFound(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewLessonRepository(db)

	request := validLessonRepositoryRequest(999, "Linear equations", 1)

	_, err := repo.CreateLesson(ctx, request)

	require.ErrorIs(t, err, repository.ErrCourseNotFound)
}

func TestLessonRepository_CreateLesson_TitleTaken(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewLessonRepository(db)
	courseID := createTestCourse(t, ctx, db, "Algebra")

	_, err := repo.CreateLesson(ctx, validLessonRepositoryRequest(courseID, "Linear equations", 1))
	require.NoError(t, err)

	_, err = repo.CreateLesson(ctx, validLessonRepositoryRequest(courseID, "Linear equations", 2))

	require.ErrorIs(t, err, repository.ErrTitleTaken)
}

func TestLessonRepository_CreateLesson_PositionTaken(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewLessonRepository(db)
	courseID := createTestCourse(t, ctx, db, "Algebra")

	_, err := repo.CreateLesson(ctx, validLessonRepositoryRequest(courseID, "Linear equations", 1))
	require.NoError(t, err)

	_, err = repo.CreateLesson(ctx, validLessonRepositoryRequest(courseID, "Quadratic equations", 1))

	require.ErrorIs(t, err, repository.ErrPositionTaken)
}

func TestLessonRepository_GetListLessonsByCourseID(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewLessonRepository(db)
	courseID := createTestCourse(t, ctx, db, "Algebra")

	lessonPosition2, err := repo.CreateLesson(ctx, validLessonRepositoryRequest(courseID, "Quadratic equations", 2))
	require.NoError(t, err)
	lessonPosition1, err := repo.CreateLesson(ctx, validLessonRepositoryRequest(courseID, "Linear equations", 1))
	require.NoError(t, err)
	lessonPosition3, err := repo.CreateLesson(ctx, validLessonRepositoryRequest(courseID, "Inequalities", 3))
	require.NoError(t, err)

	lessons, err := repo.GetListLessonsByCourseID(ctx, courseID)

	require.NoError(t, err)
	require.Len(t, lessons, 3)
	require.Equal(t, []int64{lessonPosition1.ID, lessonPosition2.ID, lessonPosition3.ID}, []int64{lessons[0].ID, lessons[1].ID, lessons[2].ID})
	require.Equal(t, []int{1, 2, 3}, []int{lessons[0].Position, lessons[1].Position, lessons[2].Position})
}

func TestLessonRepository_GetListLessonsByCourseID_Empty(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewLessonRepository(db)
	courseID := createTestCourse(t, ctx, db, "Algebra")

	lessons, err := repo.GetListLessonsByCourseID(ctx, courseID)

	require.NoError(t, err)
	require.Empty(t, lessons)
}

func TestLessonRepository_GetListLessonsByCourseID_CourseNotFound(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewLessonRepository(db)

	_, err := repo.GetListLessonsByCourseID(ctx, 999)

	require.ErrorIs(t, err, repository.ErrCourseNotFound)
}

func TestLessonRepository_GetLessonByID(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewLessonRepository(db)
	courseID := createTestCourse(t, ctx, db, "Algebra")

	created, err := repo.CreateLesson(ctx, validLessonRepositoryRequest(courseID, "Linear equations", 1))
	require.NoError(t, err)

	lesson, err := repo.GetLessonByID(ctx, created.ID)

	require.NoError(t, err)
	require.Equal(t, created, lesson)
}

func TestLessonRepository_GetLessonByID_NotFound(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewLessonRepository(db)

	_, err := repo.GetLessonByID(ctx, 999)

	require.ErrorIs(t, err, repository.ErrLessonNotFound)
}

func createTestCourse(t *testing.T, ctx context.Context, db *pgxpool.Pool, title string) int64 {
	t.Helper()

	var id int64
	err := db.QueryRow(ctx, `
		INSERT INTO courses (title, description, duration_months)
		VALUES ($1, $2, $3)
		RETURNING id
	`, title, "Course description", 3).Scan(&id)
	require.NoError(t, err)

	return id
}

func validLessonRepositoryRequest(courseID int64, title string, position int64) model.CreateLessonRequest {
	return model.CreateLessonRequest{
		CourseID:    courseID,
		Title:       title,
		Description: "Learn how to solve simple linear equations.",
		Content:     "This lesson explains equations step by step with examples.",
		Position:    position,
	}
}
