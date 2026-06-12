package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
)

type GroupRepository struct {
	pool *pgxpool.Pool
}

func NewGroupRepository(pool *pgxpool.Pool) *GroupRepository {
	return &GroupRepository{pool: pool}
}

func (r *GroupRepository) GetGroupByID(ctx context.Context, id int64) (model.Group, error) {
	query := `
		SELECT id, course_id, title, start_date, end_date, created_at, updated_at
		FROM groups
		WHERE id = $1
	`

	var group model.Group
	err := r.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&group.ID,
		&group.CourseID,
		&group.Title,
		&group.StartDate,
		&group.EndDate,
		&group.CreatedAt,
		&group.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Group{}, repository.ErrNotFound
		}
		return model.Group{}, err
	}

	return group, nil
}

func (r *GroupRepository) GroupExistsByID(ctx context.Context, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM groups
			WHERE id = $1
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
