package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
	repository_common "github.com/mathgeek-lms/mathgeek-backend/internal/repository/common"
)

type CourseRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) *CourseRepository {
	return &CourseRepository{pool: pool}
}

func (r *CourseRepository) CreateCourse(ctx context.Context, request model.CreateCourseRequest) (model.Course, error) {
	query := `
		INSERT INTO courses (title, description, duration_months)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, duration_months, created_at, updated_at
	`

	var course model.Course

	if err := r.pool.QueryRow(
		ctx,
		query,
		request.Title,
		request.Description,
		request.DurationMonths,
	).Scan(
		&course.ID,
		&course.Title,
		&course.Description,
		&course.DurationMonths,
		&course.CreatedAt,
		&course.UpdatedAt,
	); err != nil {
		if repository_common.IsPgError(err, "23505") {
			return model.Course{}, repository.ErrEmailTaken
		}
		return model.Course{}, err
	}
	return course, nil
}
