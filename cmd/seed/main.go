package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}

	dsn := os.Getenv("USERS_DB_DSN")
	if dsn == "" {
		log.Fatalf("USERS_DB_DSN is empty")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	const (
		courseTitle       = "Algebra Basics"
		courseDescription = "Introductory algebra course for development data"
		groupTitle        = "Algebra Basics Group A"
	)

	var courseID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO courses (title, description, duration_months)
		VALUES ($1, $2, $3)
		ON CONFLICT (title) DO UPDATE SET
			description = EXCLUDED.description,
			duration_months = EXCLUDED.duration_months,
			updated_at = NOW()
		RETURNING id
	`,
		courseTitle,
		courseDescription,
		3,
	).Scan(&courseID)
	if err != nil {
		log.Fatalf("failed to seed courses: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO groups (course_id, title)
		VALUES ($1, $2) 
		ON CONFLICT (course_id, title) DO UPDATE SET
			updated_at = NOW()
	`,
		courseID,
		groupTitle,
	)

	if err != nil {
		log.Fatalf("failed to seed groups: %v", err)
	}

	log.Printf("seed completed: course_id=%d, group_title=%q", courseID, groupTitle)
}
