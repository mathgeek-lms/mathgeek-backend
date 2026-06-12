package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/mathgeek-lms/mathgeek-backend/internal/handler"
	postgres "github.com/mathgeek-lms/mathgeek-backend/internal/repository/postgres"
	"github.com/mathgeek-lms/mathgeek-backend/internal/service"
	"github.com/pressly/goose/v3"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}

	dsn := buildDSN()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("pgxpool.New:", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("No connection with PostgreSQL: %v", err)
	}
	log.Println("Connection pool pgx was created")

	db, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("Error getting connection: %v", err)
	}
	db.Release()

	gooseDB, err := openGooseDB(dsn)
	if err != nil {
		log.Fatalf("Error initializing goose: %v", err)
	}
	defer gooseDB.Close()

	goose.SetDialect("postgres")
	if err := goose.Up(gooseDB, "migrations"); err != nil {
		log.Fatalf("Error while applying migrations: %v", err)
	}
	log.Println("Migrations applied")

	userRepository := postgres.NewUserRepository(pool)
	userService := service.NewUserService(userRepository)

	courseRepository := postgres.NewCourseRepository(pool)
	courseService := service.NewCourseService(courseRepository)

	groupRepository := postgres.NewGroupRepository(pool)
	groupService := service.NewGroupService(groupRepository)

	enrollmentRepository := postgres.NewEnrollmentRepository(pool)
	enrollmentService := service.NewEnrollmentService(enrollmentRepository, *groupService)

	lessonRepository := postgres.NewLessonRepository(pool)
	lessonService := service.NewLessonService(lessonRepository, enrollmentService)

	tokenSecret := os.Getenv("JWT_SECRET")
	if tokenSecret == "" {
		log.Fatalf("JWT_SECRET not found in .env. Stopping server...")
	}

	tokenService := service.NewTokenService(tokenSecret)
	router := handler.NewRouter(userService, tokenService, courseService, lessonService, groupService, enrollmentService)

	server := http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("HTTP-server started on %s", ":8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Ending work...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error while stopping: %v", err)
	}

	log.Println("Server stopped")
}

func buildDSN() string {
	return os.Getenv("USERS_DB_DSN")
}

func openGooseDB(dsn string) (*sql.DB, error) {
	return sql.Open("pgx", dsn)
}
