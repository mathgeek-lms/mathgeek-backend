package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	repository_common "github.com/mathgeek-lms/mathgeek-backend/internal/repository/common"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

func applyMigrations(t *testing.T, dsn string) {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.Up(db, "../../../migrations"))
}

func cleanupDB(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	t.Helper()

	_, err := db.Exec(ctx, `
		TRUNCATE TABLE users RESTART IDENTITY CASCADE
	`)
	require.NoError(t, err)
}

func setupTestDb(t *testing.T) (context.Context, *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	_ = godotenv.Load("../../../.env")
	dsn := os.Getenv("USERS_DB_DSN")
	if dsn == "" {
		t.Skip("USERS_DB_DSN is not set")
	}

	applyMigrations(t, dsn)

	db, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	err = db.Ping(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	cleanupDB(t, ctx, db)

	return ctx, db
}

func TestUserRepository_CreateUser(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewUserRepository(db)

	phoneNumber := "+79991234567"
	user := model.CreateUserRequest{
		Name:        "vasya",
		LastName:    "pupkin",
		Email:       "vasyapupkin777@gmail.com",
		PhoneNumber: &phoneNumber,
		Password:    "12345678",
	}

	response, err := repo.CreateUser(ctx, user)
	require.NoError(t, err)
	require.NotZero(t, response.ID)
	require.Equal(t, user.Name, response.Name)
	require.Equal(t, user.LastName, response.LastName)
	require.Equal(t, user.Email, response.Email)
	require.Equal(t, user.PhoneNumber, response.PhoneNumber)
	require.Equal(t, "STUDENT", response.Role)
	require.NotZero(t, response.CreatedAt)
	require.NotZero(t, response.UpdatedAt)

	var saved model.User
	err = db.QueryRow(ctx, `
		SELECT id, name, last_name, email, phone_number, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`, response.ID).Scan(
		&saved.ID,
		&saved.Name,
		&saved.LastName,
		&saved.Email,
		&saved.PhoneNumber,
		&saved.PasswordHash,
		&saved.Role,
		&saved.CreatedAt,
		&saved.UpdatedAt,
	)
	require.NoError(t, err)

	require.Equal(t, response.ID, saved.ID)
	require.Equal(t, user.Name, saved.Name)
	require.Equal(t, user.LastName, saved.LastName)
	require.Equal(t, user.Email, saved.Email)
	require.Equal(t, user.PhoneNumber, saved.PhoneNumber)
	require.Equal(t, "STUDENT", saved.Role)
	require.WithinDuration(t, response.CreatedAt, saved.CreatedAt, time.Second)
	require.WithinDuration(t, response.UpdatedAt, saved.UpdatedAt, time.Second)
	require.NotEqual(t, user.Password, saved.PasswordHash)
	require.True(t, repository_common.ComparePasswordHash(user.Password, saved.PasswordHash))
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewUserRepository(db)
	phoneNumber := "+79991234567"
	user := model.CreateUserRequest{
		Name:        "vasya",
		LastName:    "pupkin",
		Email:       "vasyapupkin777@gmail.com",
		PhoneNumber: &phoneNumber,
		Password:    "12345678",
	}

	created, err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	getted, err := repo.GetUserByEmail(ctx, created.Email)

	require.NoError(t, err)
	require.Equal(t, created.ID, getted.ID)
	require.Equal(t, created.Name, getted.Name)
	require.Equal(t, created.LastName, getted.LastName)
	require.Equal(t, created.Email, getted.Email)
	require.Equal(t, created.PhoneNumber, getted.PhoneNumber)
	require.NotEmpty(t, getted.PasswordHash)
	require.Equal(t, created.CreatedAt, getted.CreatedAt)
	require.Equal(t, created.UpdatedAt, getted.UpdatedAt)
}

func TestUserRepository_GetUserByID(t *testing.T) {
	ctx, db := setupTestDb(t)
	repo := NewUserRepository(db)
	phoneNumber := "+79991234567"
	user := model.CreateUserRequest{
		Name:        "vasya",
		LastName:    "pupkin",
		Email:       "vasyapupkin777@gmail.com",
		PhoneNumber: &phoneNumber,
		Password:    "12345678",
	}

	created, err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	getted, err := repo.GetUserByID(ctx, created.ID)

	require.NoError(t, err)
	require.Equal(t, created.ID, getted.ID)
	require.Equal(t, created.Name, getted.Name)
	require.Equal(t, created.LastName, getted.LastName)
	require.Equal(t, created.Email, getted.Email)
	require.Equal(t, created.PhoneNumber, getted.PhoneNumber)
	require.NotEmpty(t, getted.PasswordHash)
	require.Equal(t, created.CreatedAt, getted.CreatedAt)
	require.Equal(t, created.UpdatedAt, getted.UpdatedAt)
}
