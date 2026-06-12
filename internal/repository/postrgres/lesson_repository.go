package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
	repository_common "github.com/mathgeek-lms/mathgeek-backend/internal/repository/common"
)

type LessonRepository struct {
	pool *pgxpool.Pool
}

func NewLessonRepository(pool *pgxpool.Pool) *LessonRepository {
	return &LessonRepository{pool: pool}
}

func (r *LessonRepository) CreateLesson(ctx context.Context, request model.CreateLessonRequest) (model.Lesson, error) {
	query := `
		INSERT INTO lessons (course_id, title, description, content, position)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, course_id, title, COALESCE(description, ''), COALESCE(content, ''), position, created_at, updated_at
	`

	var lesson model.Lesson
	if err := r.pool.QueryRow(
		ctx,
		query,
		request.CourseID,
		request.Title,
		request.Description,
		request.Content,
		request.Position,
	).Scan(
		&lesson.ID,
		&lesson.CourseID,
		&lesson.Title,
		&lesson.Description,
		&lesson.Content,
		&lesson.Position,
		&lesson.CreatedAt,
		&lesson.UpdatedAt,
	); err != nil {
		if repository_common.IsPgError(err, "23503") {
			return model.Lesson{}, repository.ErrCourseNotFound
		}
		if repository_common.IsPgError(err, "23505") {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.ConstraintName {
				case "lessons_title_key":
					return model.Lesson{}, repository.ErrTitleTaken
				case "lessons_course_position_unique":
					return model.Lesson{}, repository.ErrPositionTaken
				}
			}
		}

		return model.Lesson{}, err
	}

	return lesson, nil
}

func (r *LessonRepository) GetListLessonsByCourseID(ctx context.Context, courseID int64) ([]model.Lesson, error) {
	var courseExists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM courses WHERE id = $1)`, courseID).Scan(&courseExists); err != nil {
		return nil, err
	}
	if !courseExists {
		return nil, repository.ErrCourseNotFound
	}

	query := `
		SELECT id, course_id, title, COALESCE(description, ''), COALESCE(content, ''), position, created_at, updated_at
		FROM lessons
		WHERE course_id = $1
		ORDER BY position
	`

	rows, err := r.pool.Query(ctx, query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lessons := make([]model.Lesson, 0)
	for rows.Next() {
		var lesson model.Lesson
		if err := rows.Scan(
			&lesson.ID,
			&lesson.CourseID,
			&lesson.Title,
			&lesson.Description,
			&lesson.Content,
			&lesson.Position,
			&lesson.CreatedAt,
			&lesson.UpdatedAt,
		); err != nil {
			return nil, err
		}

		lessons = append(lessons, lesson)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lessons, nil
}

func (r *LessonRepository) GetLessonByID(ctx context.Context, id int64) (model.Lesson, error) {
	query := `
		SELECT id, course_id, title, COALESCE(description, ''), COALESCE(content, ''), position, created_at, updated_at
		FROM lessons
		WHERE id = $1
	`

	var lesson model.Lesson
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&lesson.ID,
		&lesson.CourseID,
		&lesson.Title,
		&lesson.Description,
		&lesson.Content,
		&lesson.Position,
		&lesson.CreatedAt,
		&lesson.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Lesson{}, repository.ErrLessonNotFound
		}

		return model.Lesson{}, err
	}

	return lesson, nil
}
