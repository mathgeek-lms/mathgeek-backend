# Mathgeek Backend

Mathgeek Backend is the Go API for the Mathgeek LMS. It handles users, JWT auth, courses, lessons, groups, and enrollments.

This README is written for the team. It explains how to run the project locally, how auth works, and how to create the first local `ADMIN` user.

## Tech Stack

- Go
- PostgreSQL
- chi router
- JWT auth
- goose migrations
- Docker Compose for local PostgreSQL

## Local Setup

Create a `.env` file in the project root. Example:

```env
USERS_DB_NAME=mathgeek
USERS_DB_USER=mathgeek
USERS_DB_PASSWORD=mathgeek
USERS_DB_DSN=postgres://mathgeek:mathgeek@localhost:5432/mathgeek?sslmode=disable
JWT_SECRET=local-dev-secret
```

Start PostgreSQL:

```sh
docker compose up -d
```

Run migrations:

```sh
make migrate-up
```

Start the API:

```sh
go run ./cmd/app
```

The API runs on:

```text
http://localhost:8080
```

## Useful Commands

Run all tests:

```sh
go test ./...
```

Connect to the local database:

```sh
make connect-to-database
```

Create a new migration:

```sh
make migrate-create NAME=add_some_table
```

Run seed data:

```sh
go run ./cmd/seed
```

The seed command uses `USERS_DB_DSN` from `.env`. It creates or updates the `Algebra Basics` course and the `Algebra Basics Group A` group.

## Auth Model

Public registration always creates users with the `STUDENT` role.

```text
POST /api/v1/auth/register
```

Login returns a JWT access token:

```text
POST /api/v1/auth/login
```

The JWT includes the user role in its claims:

```json
{
  "user_id": 1,
  "role": "STUDENT"
}
```

Protected routes use `JWTAuth`. Admin routes also use `RequireRole("ADMIN")`.

## Creating The First Local ADMIN

We do not expose a public create-admin API in the MVP. This is intentional. A public endpoint that creates admins would be risky, so local admins are created through the database during development.

Use this flow:

1. Register a normal user.
2. Open the local database.
3. Change the user role to `ADMIN` with SQL.
4. Login again.
5. Use the new token for admin endpoints.

Register a user:

```sh
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Admin",
    "last_name": "User",
    "email": "admin@example.com",
    "password": "password123"
  }'
```

Open the local database:

```sh
make connect-to-database
```

Change the role:

```sql
UPDATE users
SET role = 'ADMIN'
WHERE email = 'admin@example.com';
```

Check it:

```sql
SELECT id, email, role
FROM users
WHERE email = 'admin@example.com';
```

Login again:

```sh
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }'
```

Copy the returned `access_token`.

To check the JWT, decode it with any local JWT tool or paste the middle token part into a base64url decoder. The payload should contain:

```json
{
  "role": "ADMIN"
}
```

Now call an admin route:

```sh
curl http://localhost:8080/api/v1/admin/test \
  -H "Authorization: Bearer <access_token>"
```

Expected result: access is allowed.

## API Overview

Public auth:

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
```

Authenticated user routes:

```text
GET  /api/v1/me
GET  /api/v1/me/enrollments
POST /api/v1/enrollments
```

Public course and lesson listing:

```text
GET /api/v1/courses/
GET /api/v1/courses/{courseID}
GET /api/v1/courses/{courseID}/lessons
```

Protected lesson access:

```text
GET /api/v1/lessons/{lessonID}
```

Admin routes:

```text
GET   /api/v1/admin/test
POST  /api/v1/admin/courses
PATCH /api/v1/admin/courses/{courseId}
POST  /api/v1/admin/lessons
PATCH /api/v1/admin/lessons/{lessonId}
POST  /api/v1/admin/groups
```

## Roles

`STUDENT`:

- Created by public registration.
- Can use normal authenticated user routes.
- Cannot use admin routes.

`ADMIN`:

- Created manually in the local database for now.
- Can use admin routes.
- Can manage courses, lessons, and groups.

## HTTP Status Rules

Use these rules when adding new endpoints:

- `401 Unauthorized`: missing token, invalid token, expired token.
- `403 Forbidden`: token is valid, but the role is not allowed.
- `400 Bad Request`: invalid JSON or invalid request fields.
- `404 Not Found`: referenced data does not exist.
- `500 Internal Server Error`: unexpected server or database error.

## Notes For Development

- Keep public registration as `STUDENT`.
- Do not add a public create-admin endpoint for MVP.
- Put admin-only routes under `/api/v1/admin`.
- Protect admin routes with both `JWTAuth` and `RequireRole("ADMIN")`.
- Prefer request validation in service code before relying on database constraints.
- Add handler tests for auth and role behavior: no token, student token, admin token.
