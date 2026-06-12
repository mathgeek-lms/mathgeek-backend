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
		log.Fatalf("USERS_DSN_DB is empty")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	// course seed
	_, err = pool.Exec(ctx, `
		INSERT INTO courses(title, description, duration_months)
		VALUES ($1, $2, $3)
		ON CONFLICT (title) DO NOTHING
	`,
		"seed course",
		"very nice and cool seed course",
		3,
	)
	if err != nil {
		log.Fatalf("failed to seed courses: %v", err)
	}

	// groups seed
	_, err = pool.Exec(ctx, `
		INSERT INTO groups (course_id, title)
		VALUES ($1, $2) 
		ON CONFLICT (title) DO NOTHING
	`,
		"1",
		"very nice and cool group",
	)

	if err != nil {
		log.Fatalf("failed to seed groups: %v", err)
	}

	log.Println("seed completed")
}
