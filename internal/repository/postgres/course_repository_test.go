package postgres

import (
	"testing"
	"time"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestCourseRepository_CreateCourse(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewCourseRepository(db)

	description := "Learn algebra from scratch."
	request := model.CreateCourseRequest{
		Title:          "Algebra",
		Description:    &description,
		DurationMonths: 3,
	}

	course, err := repo.CreateCourse(ctx, request)

	require.NoError(t, err)
	require.NotZero(t, course.ID)
	require.Equal(t, request.Title, course.Title)
	require.Equal(t, description, course.Description)
	require.Equal(t, request.DurationMonths, course.DurationMonths)
	require.NotZero(t, course.CreatedAt)
	require.NotZero(t, course.UpdatedAt)

	var saved model.Course
	err = db.QueryRow(ctx, `
		SELECT id, title, COALESCE(description, ''), duration_months, created_at, updated_at
		FROM courses
		WHERE id = $1
	`, course.ID).Scan(
		&saved.ID,
		&saved.Title,
		&saved.Description,
		&saved.DurationMonths,
		&saved.CreatedAt,
		&saved.UpdatedAt,
	)
	require.NoError(t, err)
	require.Equal(t, course.ID, saved.ID)
	require.Equal(t, request.Title, saved.Title)
	require.Equal(t, description, saved.Description)
	require.Equal(t, request.DurationMonths, saved.DurationMonths)
	require.WithinDuration(t, course.CreatedAt, saved.CreatedAt, time.Second)
	require.WithinDuration(t, course.UpdatedAt, saved.UpdatedAt, time.Second)
}

func TestCourseRepository_CreateCourse_NilDescription(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewCourseRepository(db)

	request := model.CreateCourseRequest{
		Title:          "Algebra",
		Description:    nil,
		DurationMonths: 3,
	}

	course, err := repo.CreateCourse(ctx, request)

	require.NoError(t, err)
	require.Equal(t, "", course.Description)
}

func TestCourseRepository_CreateCourse_TitleTaken(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewCourseRepository(db)

	_, err := repo.CreateCourse(ctx, validCourseRepositoryRequest("Algebra"))
	require.NoError(t, err)

	_, err = repo.CreateCourse(ctx, validCourseRepositoryRequest("Algebra"))

	require.ErrorIs(t, err, repository.ErrTitleTaken)
}

func TestCourseRepository_GetListCourses(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewCourseRepository(db)

	firstCourse, err := repo.CreateCourse(ctx, validCourseRepositoryRequest("Algebra"))
	require.NoError(t, err)
	secondCourse, err := repo.CreateCourse(ctx, validCourseRepositoryRequest("Geometry"))
	require.NoError(t, err)

	courses, err := repo.GetListCourses(ctx)

	require.NoError(t, err)
	require.Len(t, courses, 2)
	require.Equal(t, []int64{firstCourse.ID, secondCourse.ID}, []int64{courses[0].ID, courses[1].ID})
	require.Equal(t, []string{firstCourse.Title, secondCourse.Title}, []string{courses[0].Title, courses[1].Title})
}

func TestCourseRepository_GetListCourses_Empty(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewCourseRepository(db)

	courses, err := repo.GetListCourses(ctx)

	require.NoError(t, err)
	require.Empty(t, courses)
}

func TestCourseRepository_GetCourseByID(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewCourseRepository(db)

	created, err := repo.CreateCourse(ctx, validCourseRepositoryRequest("Algebra"))
	require.NoError(t, err)

	course, err := repo.GetCourseByID(ctx, created.ID)

	require.NoError(t, err)
	require.Equal(t, created, course)
}

func TestCourseRepository_GetCourseByID_NilDescription(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewCourseRepository(db)

	created, err := repo.CreateCourse(ctx, model.CreateCourseRequest{
		Title:          "Algebra",
		Description:    nil,
		DurationMonths: 3,
	})
	require.NoError(t, err)

	course, err := repo.GetCourseByID(ctx, created.ID)

	require.NoError(t, err)
	require.Equal(t, "", course.Description)
}

func TestCourseRepository_GetCourseByID_NotFound(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewCourseRepository(db)

	_, err := repo.GetCourseByID(ctx, 999)

	require.ErrorIs(t, err, repository.ErrNotFound)
}

func validCourseRepositoryRequest(title string) model.CreateCourseRequest {
	description := "Learn math from scratch."

	return model.CreateCourseRequest{
		Title:          title,
		Description:    &description,
		DurationMonths: 3,
	}
}
