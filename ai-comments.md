# Sprint 1 - until 14.06.2026 23:59\

## Tasks for user 1 (undndnwnkk on github)
### MG-3:
Goal:
Create the minimal Go backend application so we have something that starts and responds to HTTP requests.

What to do:

create Go module with go mod init;

create cmd/api/main.go;

start HTTP server on port 8080;

add endpoint GET /health;

return JSON response {"status":"ok"};

add README with run instructions.

How to check:

Run the app:
go run ./cmd/api

In another terminal run:
curl http://localhost:8080/health

Expected result:
{"status":"ok"}

What to read/watch if stuck:

Go official tutorial: Create a Go module

Go net/http package basics

Article/video: building a simple HTTP server in Go

Notes for beginners:
Do not add database or auth in this task. Keep it tiny. The only goal is to make the backend start and answer /health.

### MG-4:
Goal:
Prepare local database setup for the backend.

What to do:

add docker-compose.yml with PostgreSQL;

add DATABASE_URL env variable example;

connect backend to Postgres on startup;

add migrations folder;

add Makefile commands migrate-up and migrate-down;

create first test migration if needed.

How to check:

Run:
docker compose up -d

Run backend:
go run ./cmd/api

Backend should start without database connection errors.

Run migrations:
make migrate-up

What to read/watch if stuck:

Docker Compose basics

PostgreSQL connection string format

golang-migrate getting started

Go database/sql package basics

Notes for beginners:
Do not build users yet. This task is only about making the database available and migration flow repeatable.

### MG-5:
Goal:
Create the database table and Go code needed to store and read users.

What to do:

create migration for users table;

fields: id, name, last_name, email, phone_number, password_hash, role, created_at, updated_at;

make email unique;

create User struct in Go;

create repository methods: CreateUser, GetUserByEmail, GetUserByID.

How to check:

Run make migrate-up.

Check table exists in Postgres.

Write a small temporary test or use repository from handler later.

Try inserting two users with the same email and confirm the second insert fails.

What to read/watch if stuck:

SQL CREATE TABLE basics

PostgreSQL UNIQUE constraint

Go structs

pgx or database/sql examples

Notes for beginners:
The repository should not know anything about HTTP. It only talks to the database.

### MG-6:
Goal:
Allow a new student to create an account.

What to do:

create request struct with email, password, name, last_name;

validate required fields;

hash password with bcrypt;

save user to database with role student;

return user data without password_hash;

handle duplicate email with clear error.

Endpoint:
POST /auth/register

Example request fields:
email, password, name, last_name

How to check:

Start backend and database.

Send a POST request to /auth/register with valid JSON.

Check response contains id, email, name, last_name, role.

Check response does not contain password or password_hash.

Try the same email twice. Second request should fail with 409 or a clear validation error.

What to read/watch if stuck:

Go JSON decoding with json.NewDecoder

bcrypt in Go: golang.org/x/crypto/bcrypt

HTTP status codes 201, 400, 409

Notes for beginners:
Never store raw passwords. Always store only password_hash.

### MG-7:
Goal:
Allow an existing user to log in and receive a JWT token.

What to do:

create login request with email and password;

find user by email;

compare provided password with password_hash using bcrypt;

return unauthorized error for wrong credentials;

generate JWT containing user_id and role;

return token and basic user info.

Endpoint:
POST /auth/login

How to check:

Register a user first.

Send login request with the same email and password.

Check response contains token and user object.

Try wrong password. It should return 401.

Decode token locally and check it contains user_id and role.

What to read/watch if stuck:

JWT basics

Go JWT library examples

bcrypt CompareHashAndPassword

HTTP 401 Unauthorized

Notes for beginners:
Do not reveal whether email or password is wrong. Return a generic login error.

### MG-8:
Goal:
Protect API endpoints with JWT and allow the current user to read their own profile.

What to do:

create middleware that reads Authorization header;

validate Bearer token;

put user_id and role into request context;

create GET /me endpoint;

load current user from database;

return profile without password_hash.

Endpoints:
GET /me

How to check:

Call GET /me without token. Expected: 401.

Register and login to get token.

Call GET /me with Authorization: Bearer token.

Expected: current user profile.

Check password_hash is not returned.

What to read/watch if stuck:

Go HTTP middleware pattern

context.Context basics

Authorization header and Bearer tokens

JWT validation examples in Go

Notes for beginners:
Middleware is just a function that runs before the handler. Keep it simple.

## Task for user 2(absoluteqq on github)
### MG-9:
Goal:
Prepare database schema for course and lesson data.

What to do:

create courses table;

create lessons table;

connect lessons to courses with course_id foreign key;

add position field to lessons;

add created_at and updated_at fields.

Suggested fields for courses:
id, title, description, duration_months, created_at, updated_at

Suggested fields for lessons:
id, course_id, title, description, content, position, created_at, updated_at

How to check:

Run make migrate-up.

Open Postgres and check both tables exist.

Insert a course manually.

Insert a lesson connected to that course.

Try inserting lesson with fake course_id. It should fail because of foreign key.

What to read/watch if stuck:

SQL foreign keys

PostgreSQL UUID columns

Database migrations with golang-migrate

Basic relational database design

Notes for beginners:
A lesson belongs to one course. That is why lessons.course_id references courses.id.

### MG-10:
Goal:
Make local development easier by adding test course data.

What to do:

add one test course, for example Algebra Basics;

add three lessons for this course;

make sure lesson positions are 1, 2, 3;

document how to load seed data.

How to check:

Run migrations.

Run seed script or seed SQL.

Open database and check there is one course.

Check the course has three lessons.

Check lessons are ordered by position.

What to read/watch if stuck:

SQL INSERT statements

psql basics

Docker exec into PostgreSQL container

Difference between migrations and seed data

Notes for beginners:
Seed data is fake data for development. It helps test API without building admin features first.

### MG-11:
Goal:
Create Go code that reads courses from PostgreSQL.

What to do:

create Course struct;

create CourseRepository;

implement GetListCourses;

implement GetCourseByID;

return clear not found error when course does not exist.

How to check:

Load seed data.

Call ListCourses from a temporary test or handler.

Check it returns the test course.

Call GetCourseByID with existing id. Expected: course returned.

Call GetCourseByID with random id. Expected: not found error.

What to read/watch if stuck:

Go methods and structs

SQL SELECT queries

pgx Query and QueryRow examples

Handling sql.ErrNoRows or pgx.ErrNoRows

Notes for beginners:
Repository should return Go data, not HTTP responses. HTTP status codes belong in handlers.

### MG-12:
Goal:
Expose course data through HTTP API.

What to do:

create course HTTP handlers;

implement GET /courses;

implement GET /courses/{courseId};

return JSON responses;

return 404 when course does not exist;

keep response fields simple: id, title, description, duration_months.

How to check:

Start backend and database.

Load seed data.

Call GET /courses. Expected: list with test course.

Copy course id and call GET /courses/{courseId}. Expected: one course.

Call GET /courses with random id. Expected: 404.

What to read/watch if stuck:

chi or gin route parameters

JSON encoding with json.NewEncoder

HTTP 404 Not Found

curl basics

Notes for beginners:
At this stage endpoints can be public. Later we will protect them with auth middleware.

### MG-13:
Goal:
Create Go code that reads lessons from PostgreSQL.

What to do:

create Lesson struct;

create LessonRepository;

implement ListLessonsByCourseID;

implement GetLessonByID;

order lessons by position;

return clear not found error when lesson does not exist.

How to check:

Load seed data.

Call ListLessonsByCourseID with test course id.

Expected: three lessons ordered by position.

Call GetLessonByID with existing lesson id. Expected: lesson returned.

Call GetLessonByID with random id. Expected: not found error.

What to read/watch if stuck:

SQL WHERE and ORDER BY

Go slices

pgx rows.Next loop

Error handling in Go

Notes for beginners:
Start with simple SELECT queries. Do not add filtering, pagination, or permissions in this task.

### MG-14:
Goal:
Expose lesson data through HTTP API.

What to do:

implement GET /courses/{courseId}/lessons;

implement GET /lessons/{lessonId};

return lessons ordered by position;

return lesson content only on lesson detail endpoint;

return 404 when course or lesson does not exist.

How to check:

Start backend and database.

Load seed data.

Call GET /courses/{courseId}/lessons. Expected: list of lessons.

Check lessons are ordered by position.

Copy lesson id and call GET /lessons/{lessonId}. Expected: lesson detail with content.

Try random lesson id. Expected: 404.

What to read/watch if stuck:

Nested routes in chi or gin

URL parameters in Go routers

JSON response design

HTTP 404 handling

Notes for beginners:
List endpoint should return short lesson info. Detail endpoint can return full content.