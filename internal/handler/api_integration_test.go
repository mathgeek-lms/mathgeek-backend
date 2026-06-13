package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	pgrepository "github.com/mathgeek-lms/mathgeek-backend/internal/repository/postgres"
	"github.com/mathgeek-lms/mathgeek-backend/internal/service"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type apiIntegrationApp struct {
	ctx    context.Context
	db     *pgxpool.Pool
	router http.Handler
}

type apiIntegrationFixture struct {
	studentToken string
	adminToken   string
	otherToken   string
	studentID    int64
	courseID     int64
	lessonID     int64
	groupID      int64
}

func TestAPIIntegration_MainBackendFlow(t *testing.T) {
	app := setupAPIIntegrationApp(t)

	student := app.registerUser(t, "Student", "User", "student@example.com", "password123")
	studentID := int64Field(t, student, "id")
	require.Equal(t, "STUDENT", stringField(t, student, "role"))
	require.NotContains(t, student, "password")
	require.NotContains(t, student, "password_hash")

	studentToken := app.loginUser(t, "student@example.com", "password123")

	meResponse := app.doJSON(t, http.MethodGet, "/api/v1/me", studentToken, "")
	require.Equal(t, http.StatusOK, meResponse.Code)

	var me map[string]any
	decodeResponse(t, meResponse, &me)
	require.Equal(t, studentID, int64Field(t, me, "id"))
	require.Equal(t, "student@example.com", stringField(t, me, "email"))
	require.Equal(t, "STUDENT", stringField(t, me, "role"))

	adminToken := app.createAdminToken(t)
	courseID := app.createCourse(t, adminToken)
	lessonID := app.createLesson(t, adminToken, courseID)
	groupID := app.createGroup(t, adminToken, courseID)

	coursesResponse := app.doJSON(t, http.MethodGet, "/api/v1/courses/", "", "")
	require.Equal(t, http.StatusOK, coursesResponse.Code)

	var courses []map[string]any
	decodeResponse(t, coursesResponse, &courses)
	require.Len(t, courses, 1)
	require.Equal(t, courseID, int64Field(t, courses[0], "id"))
	require.Equal(t, "Algebra", stringField(t, courses[0], "title"))

	courseResponse := app.doJSON(t, http.MethodGet, fmt.Sprintf("/api/v1/courses/%d", courseID), "", "")
	require.Equal(t, http.StatusOK, courseResponse.Code)

	var course map[string]any
	decodeResponse(t, courseResponse, &course)
	require.Equal(t, courseID, int64Field(t, course, "id"))
	require.Equal(t, "Algebra", stringField(t, course, "title"))

	lessonsResponse := app.doJSON(t, http.MethodGet, fmt.Sprintf("/api/v1/courses/%d/lessons", courseID), "", "")
	require.Equal(t, http.StatusOK, lessonsResponse.Code)

	var lessons []map[string]any
	decodeResponse(t, lessonsResponse, &lessons)
	require.Len(t, lessons, 1)
	require.Equal(t, lessonID, int64Field(t, lessons[0], "id"))
	require.Equal(t, "Linear equations", stringField(t, lessons[0], "title"))
	require.NotContains(t, lessons[0], "content")

	lockedLessonResponse := app.doJSON(t, http.MethodGet, fmt.Sprintf("/api/v1/lessons/%d", lessonID), studentToken, "")
	require.Equal(t, http.StatusForbidden, lockedLessonResponse.Code)
	requireAPIError(t, lockedLessonResponse, "forbidden", service.ErrNotEnrolled.Error())

	enrollmentResponse := app.doJSON(t, http.MethodPost, "/api/v1/enrollments", studentToken, fmt.Sprintf(`{"group_id":%d}`, groupID))
	require.Equal(t, http.StatusCreated, enrollmentResponse.Code)

	var enrollment map[string]any
	decodeResponse(t, enrollmentResponse, &enrollment)
	require.Equal(t, studentID, int64Field(t, enrollment, "user_id"))
	require.Equal(t, groupID, int64Field(t, enrollment, "group_id"))
	require.Equal(t, "ACTIVE", stringField(t, enrollment, "status"))

	enrollmentsResponse := app.doJSON(t, http.MethodGet, "/api/v1/me/enrollments", studentToken, "")
	require.Equal(t, http.StatusOK, enrollmentsResponse.Code)

	var enrollments []map[string]any
	decodeResponse(t, enrollmentsResponse, &enrollments)
	require.Len(t, enrollments, 1)
	require.Equal(t, groupID, int64Field(t, enrollments[0], "group_id"))
	require.Equal(t, courseID, int64Field(t, enrollments[0], "course_id"))
	require.Equal(t, "Algebra Group A", stringField(t, enrollments[0], "group_title"))
	require.Equal(t, "Algebra", stringField(t, enrollments[0], "course_title"))

	unlockedLessonResponse := app.doJSON(t, http.MethodGet, fmt.Sprintf("/api/v1/lessons/%d", lessonID), studentToken, "")
	require.Equal(t, http.StatusOK, unlockedLessonResponse.Code)

	var unlockedLesson map[string]any
	decodeResponse(t, unlockedLessonResponse, &unlockedLesson)
	require.Equal(t, lessonID, int64Field(t, unlockedLesson, "id"))
	require.Equal(t, "Full lesson content for integration tests.", stringField(t, unlockedLesson, "content"))
}

func TestAPIIntegration_FailedAPIBehavior(t *testing.T) {
	app := setupAPIIntegrationApp(t)
	fixture := app.createFixture(t)

	firstEnrollment := app.doJSON(t, http.MethodPost, "/api/v1/enrollments", fixture.studentToken, fmt.Sprintf(`{"group_id":%d}`, fixture.groupID))
	require.Equal(t, http.StatusCreated, firstEnrollment.Code)

	tests := []struct {
		name        string
		method      string
		path        string
		token       string
		body        string
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name:        "register validation error",
			method:      http.MethodPost,
			path:        "/api/v1/auth/register",
			body:        `{"name":"","last_name":"User","email":"new@example.com","password":"password123"}`,
			wantStatus:  http.StatusBadRequest,
			wantCode:    "bad_request",
			wantMessage: service.ErrEmptyName.Error(),
		},
		{
			name:        "duplicate registration",
			method:      http.MethodPost,
			path:        "/api/v1/auth/register",
			body:        `{"name":"Student","last_name":"User","email":"student@example.com","password":"password123"}`,
			wantStatus:  http.StatusConflict,
			wantCode:    "conflict",
			wantMessage: service.ErrEmailAlreadyTaken.Error(),
		},
		{
			name:        "bad login password",
			method:      http.MethodPost,
			path:        "/api/v1/auth/login",
			body:        `{"email":"student@example.com","password":"bad-password"}`,
			wantStatus:  http.StatusUnauthorized,
			wantCode:    "unauthorized",
			wantMessage: service.ErrIncorrectPassword.Error(),
		},
		{
			name:        "current user without token",
			method:      http.MethodGet,
			path:        "/api/v1/me",
			wantStatus:  http.StatusUnauthorized,
			wantCode:    "unauthorized",
			wantMessage: "missing authorization header",
		},
		{
			name:        "student cannot use admin endpoint",
			method:      http.MethodPost,
			path:        "/api/v1/admin/courses",
			token:       fixture.studentToken,
			body:        `{"title":"Geometry","description":"Learn geometry.","duration_months":2}`,
			wantStatus:  http.StatusForbidden,
			wantCode:    "forbidden",
			wantMessage: "forbidden",
		},
		{
			name:        "missing course",
			method:      http.MethodGet,
			path:        "/api/v1/courses/999",
			wantStatus:  http.StatusNotFound,
			wantCode:    "not_found",
			wantMessage: service.ErrCourseNotFound.Error(),
		},
		{
			name:        "missing lessons course",
			method:      http.MethodGet,
			path:        "/api/v1/courses/999/lessons",
			wantStatus:  http.StatusNotFound,
			wantCode:    "not_found",
			wantMessage: service.ErrCourseNotFound.Error(),
		},
		{
			name:        "unenrolled student cannot open lesson",
			method:      http.MethodGet,
			path:        fmt.Sprintf("/api/v1/lessons/%d", fixture.lessonID),
			token:       fixture.otherToken,
			wantStatus:  http.StatusForbidden,
			wantCode:    "forbidden",
			wantMessage: service.ErrNotEnrolled.Error(),
		},
		{
			name:        "enrollment missing group",
			method:      http.MethodPost,
			path:        "/api/v1/enrollments",
			token:       fixture.otherToken,
			body:        `{"group_id":999}`,
			wantStatus:  http.StatusNotFound,
			wantCode:    "not_found",
			wantMessage: service.ErrGroupNotFound.Error(),
		},
		{
			name:        "duplicate enrollment",
			method:      http.MethodPost,
			path:        "/api/v1/enrollments",
			token:       fixture.studentToken,
			body:        fmt.Sprintf(`{"group_id":%d}`, fixture.groupID),
			wantStatus:  http.StatusConflict,
			wantCode:    "conflict",
			wantMessage: service.ErrEnrollmentAlreadyExist.Error(),
		},
		{
			name:        "admin cannot enroll",
			method:      http.MethodPost,
			path:        "/api/v1/enrollments",
			token:       fixture.adminToken,
			body:        fmt.Sprintf(`{"group_id":%d}`, fixture.groupID),
			wantStatus:  http.StatusForbidden,
			wantCode:    "forbidden",
			wantMessage: "only student can enroll to group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := app.doJSON(t, tt.method, tt.path, tt.token, tt.body)

			require.Equal(t, tt.wantStatus, response.Code)
			requireAPIError(t, response, tt.wantCode, tt.wantMessage)
		})
	}
}

func setupAPIIntegrationApp(t *testing.T) *apiIntegrationApp {
	t.Helper()

	ctx := context.Background()
	testcontainers.SkipIfProviderIsNotHealthy(t)

	container, err := tcpostgres.Run(
		ctx,
		"postgres:17-alpine",
		tcpostgres.WithDatabase("mathgeek_test"),
		tcpostgres.WithUsername("mathgeek_test"),
		tcpostgres.WithPassword("mathgeek_test"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testcontainers.TerminateContainer(container))
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	applyAPIIntegrationMigrations(t, dsn)

	db, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	require.NoError(t, db.Ping(ctx))

	app := &apiIntegrationApp{
		ctx:    ctx,
		db:     db,
		router: newAPIIntegrationRouter(db),
	}

	app.cleanData(t)
	t.Cleanup(func() {
		app.cleanData(t)
		db.Close()
	})

	return app
}

func newAPIIntegrationRouter(db *pgxpool.Pool) http.Handler {
	userRepository := pgrepository.NewUserRepository(db)
	courseRepository := pgrepository.NewCourseRepository(db)
	lessonRepository := pgrepository.NewLessonRepository(db)
	groupRepository := pgrepository.NewGroupRepository(db)
	enrollmentRepository := pgrepository.NewEnrollmentRepository(db)

	userService := service.NewUserService(userRepository)
	tokenService := service.NewTokenService("integration-test-secret")
	courseService := service.NewCourseService(courseRepository)
	groupService := service.NewGroupService(groupRepository, courseService)
	enrollmentService := service.NewEnrollmentService(enrollmentRepository, *groupService)
	lessonService := service.NewLessonService(lessonRepository, enrollmentService, courseService)

	return NewRouter(userService, tokenService, courseService, lessonService, groupService, enrollmentService)
}

func applyAPIIntegrationMigrations(t *testing.T, dsn string) {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.Up(db, "../../migrations"))
}

func (a *apiIntegrationApp) cleanData(t *testing.T) {
	t.Helper()

	_, err := a.db.Exec(a.ctx, `
		TRUNCATE TABLE enrollments, groups, lessons, courses, users RESTART IDENTITY CASCADE
	`)
	require.NoError(t, err)
}

func (a *apiIntegrationApp) createFixture(t *testing.T) apiIntegrationFixture {
	t.Helper()

	student := a.registerUser(t, "Student", "User", "student@example.com", "password123")
	studentToken := a.loginUser(t, "student@example.com", "password123")

	a.registerUser(t, "Other", "Student", "other@example.com", "password123")
	otherToken := a.loginUser(t, "other@example.com", "password123")

	adminToken := a.createAdminToken(t)
	courseID := a.createCourse(t, adminToken)
	lessonID := a.createLesson(t, adminToken, courseID)
	groupID := a.createGroup(t, adminToken, courseID)

	return apiIntegrationFixture{
		studentToken: studentToken,
		adminToken:   adminToken,
		otherToken:   otherToken,
		studentID:    int64Field(t, student, "id"),
		courseID:     courseID,
		lessonID:     lessonID,
		groupID:      groupID,
	}
}

func (a *apiIntegrationApp) createAdminToken(t *testing.T) string {
	t.Helper()

	a.registerUser(t, "Admin", "User", "admin@example.com", "password123")

	tag, err := a.db.Exec(a.ctx, `
		UPDATE users
		SET role = 'ADMIN'
		WHERE email = 'admin@example.com'
	`)
	require.NoError(t, err)
	require.Equal(t, int64(1), tag.RowsAffected())

	return a.loginUser(t, "admin@example.com", "password123")
}

func (a *apiIntegrationApp) registerUser(t *testing.T, name, lastName, email, password string) map[string]any {
	t.Helper()

	body := fmt.Sprintf(
		`{"name":%q,"last_name":%q,"email":%q,"password":%q}`,
		name,
		lastName,
		email,
		password,
	)
	response := a.doJSON(t, http.MethodPost, "/api/v1/auth/register", "", body)
	require.Equal(t, http.StatusCreated, response.Code)

	var user map[string]any
	decodeResponse(t, response, &user)

	return user
}

func (a *apiIntegrationApp) loginUser(t *testing.T, email, password string) string {
	t.Helper()

	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)
	response := a.doJSON(t, http.MethodPost, "/api/v1/auth/login", "", body)
	require.Equal(t, http.StatusOK, response.Code)

	var token map[string]any
	decodeResponse(t, response, &token)
	return stringField(t, token, "access_token")
}

func (a *apiIntegrationApp) createCourse(t *testing.T, adminToken string) int64 {
	t.Helper()

	response := a.doJSON(t, http.MethodPost, "/api/v1/admin/courses", adminToken, `{
		"title":"Algebra",
		"description":"Learn algebra from scratch.",
		"duration_months":3
	}`)
	require.Equal(t, http.StatusCreated, response.Code)

	var course map[string]any
	decodeResponse(t, response, &course)
	return int64Field(t, course, "id")
}

func (a *apiIntegrationApp) createLesson(t *testing.T, adminToken string, courseID int64) int64 {
	t.Helper()

	response := a.doJSON(t, http.MethodPost, "/api/v1/admin/lessons", adminToken, fmt.Sprintf(`{
		"course_id":%d,
		"title":"Linear equations",
		"description":"Learn linear equations.",
		"content":"Full lesson content for integration tests.",
		"position":1
	}`, courseID))
	require.Equal(t, http.StatusCreated, response.Code)

	var lesson map[string]any
	decodeResponse(t, response, &lesson)
	return int64Field(t, lesson, "id")
}

func (a *apiIntegrationApp) createGroup(t *testing.T, adminToken string, courseID int64) int64 {
	t.Helper()

	response := a.doJSON(t, http.MethodPost, "/api/v1/admin/groups", adminToken, fmt.Sprintf(`{
		"course_id":%d,
		"title":"Algebra Group A"
	}`, courseID))
	require.Equal(t, http.StatusCreated, response.Code)

	var group map[string]any
	decodeResponse(t, response, &group)
	return int64Field(t, group, "id")
}

func (a *apiIntegrationApp) doJSON(t *testing.T, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	request := httptest.NewRequest(method, path, reader)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	recorder := httptest.NewRecorder()
	a.router.ServeHTTP(recorder, request)

	return recorder
}

func decodeResponse(t *testing.T, recorder *httptest.ResponseRecorder, dest any) {
	t.Helper()

	require.NoError(t, json.NewDecoder(recorder.Body).Decode(dest))
}

func int64Field(t *testing.T, data map[string]any, key string) int64 {
	t.Helper()

	value, ok := data[key].(float64)
	require.Truef(t, ok, "expected numeric %q field in %#v", key, data)

	return int64(value)
}

func stringField(t *testing.T, data map[string]any, key string) string {
	t.Helper()

	value, ok := data[key].(string)
	require.Truef(t, ok, "expected string %q field in %#v", key, data)

	return value
}
