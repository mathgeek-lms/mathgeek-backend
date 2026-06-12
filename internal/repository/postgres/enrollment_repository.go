package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
	repository_common "github.com/mathgeek-lms/mathgeek-backend/internal/repository/common"
)

type EnrollmentRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) *EnrollmentRepository {
	return &EnrollmentRepository{pool: pool}
}

func (r *EnrollmentRepository) CreateEnrollment(ctx context.Context, userID, groupID int64) (model.Enrollment, error) {
	query := `
		INSERT INTO enrollments (user_id, group_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, group_id, status, created_at, updated_at
	`

	var enrollment model.Enrollment
	err := r.pool.QueryRow(
		ctx,
		query,
		userID,
		groupID,
		"ACTIVE",
	).Scan(
		&enrollment.ID,
		&enrollment.UserID,
		&enrollment.GroupID,
		&enrollment.Status,
		&enrollment.CreatedAt,
		&enrollment.UpdatedAt,
	)

	if err != nil {
		if repository_common.IsPgError(err, "23505") {
			return model.Enrollment{}, repository.ErrEnrollmentAlreadyExists
		}

		return model.Enrollment{}, err
	}

	return enrollment, nil
}

func (r *EnrollmentRepository) ListEnrollmentsByUserID(ctx context.Context, userID int64) ([]model.EnrollmentWithDetails, error) {
	query := `
		SELECT e.id, e.status, g.id, g.title, c.id, c.title
		FROM enrollments e
		INNER JOIN groups g
			ON e.group_id = g.id
		INNER JOIN courses c
			ON g.course_id = c.id
		WHERE e.user_id = $1
	`

	list := make([]model.EnrollmentWithDetails, 0)
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var enrollment model.EnrollmentWithDetails
		if err := rows.Scan(
			&enrollment.ID,
			&enrollment.Status,
			&enrollment.GroupID,
			&enrollment.GroupTitle,
			&enrollment.CourseID,
			&enrollment.CourseTitle,
		); err != nil {
			return nil, err
		}

		list = append(list, enrollment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *EnrollmentRepository) IsUserEnrolledInCourse(ctx context.Context, userID, courseID int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM enrollments e
			INNER JOIN groups g
				ON e.group_id = g.id
			WHERE e.user_id = $1
				AND g.course_id = $2
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, courseID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
