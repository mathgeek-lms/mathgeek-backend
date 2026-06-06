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

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) CreateUser(ctx context.Context, request model.CreateUserRequest) (model.User, error) {
	query := `
			INSERT INTO users (name, last_name, email, phone_number, password_hash, role)
			VALUES ($1, $2, $3, $4, $5, 'STUDENT')
			RETURNING id, name, last_name, email, phone_number, password_hash, role, created_at, updated_at
		`

	var user model.User
	passwordHash, err := repository_common.HashPassword(request.Password)
	if err != nil {
		return model.User{}, err
	}

	err = r.pool.QueryRow(
		ctx,
		query,
		request.Name,
		request.LastName,
		request.Email,
		request.PhoneNumber,
		passwordHash,
	).Scan(
		&user.ID,
		&user.Name,
		&user.LastName,
		&user.Email,
		&user.PhoneNumber,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if isPgError(err, "23505") {
		return model.User{}, repository.ErrEmailTaken
	}
	return user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	query := `
		SELECT * FROM users 
		WHERE email = $1
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.LastName,
		&user.Email,
		&user.PhoneNumber,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, repository.ErrNotFound
		}
		return model.User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (model.User, error) {
	query := `
		SELECT * FROM users
		WHERE id = $1
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.LastName,
		&user.Email,
		&user.PhoneNumber,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, repository.ErrNotFound
		}

		return model.User{}, err
	}

	return user, nil
}

func isPgError(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}
