package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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
		RETURNING id, title, COALESCE(description, ''), duration_months, created_at, updated_at
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
			return model.Course{}, repository.ErrTitleTaken
		}
		return model.Course{}, err
	}
	return course, nil
}

func (r *CourseRepository) GetListCourses(ctx context.Context) ([]model.Course, error) {
	query := `
		SELECT id, title, COALESCE(description, ''), duration_months, created_at, updated_at
		FROM courses
		ORDER BY id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := make([]model.Course, 0)
	for rows.Next() {
		var course model.Course
		if err := rows.Scan(
			&course.ID,
			&course.Title,
			&course.Description,
			&course.DurationMonths,
			&course.CreatedAt,
			&course.UpdatedAt,
		); err != nil {
			return nil, err
		}

		courses = append(courses, course)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courses, nil
}

func (r *CourseRepository) GetCourseByID(ctx context.Context, id int64) (model.Course, error) {
	query := `
		SELECT id, title, description, duration_months, created_at, updated_at
		FROM courses
		WHERE id = $1 
	`

	var course model.Course
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&course.ID,
		&course.Title,
		&course.Description,
		&course.DurationMonths,
		&course.CreatedAt,
		&course.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, repository.ErrNotFound
		}

		return model.Course{}, err
	}

	return course, nil
}
