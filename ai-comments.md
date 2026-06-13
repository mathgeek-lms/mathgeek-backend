# Sprint 1 - until 22.06.2026 23:59\

## Tasks for user 1 (undndnwnkk on github)
MG-25:
Goal:
Add course groups to the database.

Context:
Use the same database as users and courses. Do not create a separate database. Use BIGSERIAL/BIGINT ids to match current users implementation.

What to do:

create a goose migration for groups;

add id BIGSERIAL primary key;

add course_id BIGINT not null;

reference courses(id) with foreign key;

add title;

add optional start_date and end_date;

add created_at and updated_at;

add Down migration.

Suggested table:
groups: id, course_id, title, start_date, end_date, created_at, updated_at

MG-26:
Goal:
Store which user is enrolled into which group.

What to do:

create goose migration;

create enrollments table;

fields: id, user_id, group_id, status, created_at, updated_at;

user_id references users(id);

group_id references groups(id);

status default should be ACTIVE;

add UNIQUE(user_id, group_id);

add Down migration.

How to check:

make migrate-up works;

duplicate user_id/group_id insert fails;

fake user_id fails;

fake group_id

MG-27:
Goal: add one development group for the seeded course.

Do:

add seed data for one group linked to the existing test course;

use title like Algebra Basics Group A;

document how to run seed data.

Check:

course exists;

group exists;

group.course_id points to real course;

group can be used later by POST /api/v1/enrollments.

Read:

SQL INSERT;

foreign keys;

difference between migrations and seed data.


MG-49:
Goal:
Add minimal Go code for reading groups from database. We do not expose group endpoints yet.

What to do:

create Group model;

create GroupRepository or add group read methods near enrollment repository;

implement GetGroupByID;

implement GroupExists if needed;

return repository.ErrNotFound when group does not exist.

How to check:

seed one group;

GetGroupByID(existing id) returns group;

GetGroupByID(fake id) returns ErrNotFound;

POST /api/v1/enrollments with fake group_id returns 404, not 500.

MG-28:
Goal: add database access for enrollments.

Do:

create Enrollment model;

create repository interface;

implement CreateEnrollment;

implement ListEnrollmentsByUserID;

implement IsUserEnrolledInCourse;

map duplicate enrollment to conflict error.

Check:

create enrollment works;

duplicate returns known error;

list by user returns group and course info;

IsUserEnrolledInCourse returns true/false.

Read:

SQL JOIN;

UNIQUE constraint errors;

Go repository pattern.


MG-30:
Goal: allow authenticated STUDENT to enroll into a group.

Do:

read current user_id from JWT claims;

accept group_id in request body;

validate group exists;

create enrollment with status ACTIVE;

return 201 on success;

return 409 for duplicate enrollment;

return 401 without token.

Check:

register and login user;

call POST /api/v1/enrollments;

duplicate request returns conflict;

fake group_id returns not found.

Read:

chi handlers;

request context;

HTTP 201, 401, 404, 409.



MG-31:
Goal: show current user's enrollments.

Do:

protect endpoint with JWTAuth;

read user_id from claims;

return enrollments with group and course info;

return empty array if user has no enrollments.

Response should include:

enrollment id;

status;

group id and title;

course id and title.

Check:

no token returns 401;

new user gets empty list;

enrolled user gets one enrollment;

response does not expose password data.


MG-32:
Goal: STUDENT can open lesson only if enrolled into a group of that lesson's course.

Do:

add service check for lesson access;

use current user_id from JWT claims;

find lesson course_id;

check enrollment for user and course;

return 403 when user is not enrolled;

keep ADMIN access open later.

Check:

enrolled user can GET lesson;

non-enrolled user gets 403;

no token gets 401.

Read:

authorization vs authentication;

SQL EXISTS;

service-layer access checks.

### MG-39
Goal: prepare protected admin API area.

Do:

use existing RequireRole or fix it if needed;

create /api/v1/admin route group;

protect it with JWTAuth;

require role ADMIN;

keep STUDENT blocked.

Check:

no token gets 401;

STUDENT token gets 403;

ADMIN token can access test admin endpoint.

Read:

middleware chaining in chi;

authentication vs authorization;

HTTP 401 vs 403.

### MG-40
Goal: allow ADMIN to create courses through API.

Do:

protect endpoint with JWTAuth and ADMIN role;

accept title, description, duration_months;

validate title is not empty;

validate duration_months is greater than zero;

save course to database;

return created course.

Check:

STUDENT gets 403;

ADMIN creates course;

empty title gets 400;

bad duration gets 400.

Read:

request validation;

INSERT RETURNING;

admin route groups.

### MG-41
Goal: allow ADMIN to update course data.

Do:

protect endpoint with JWTAuth and ADMIN role;

read courseId from URL;

accept title, description, duration_months;

validate provided fields;

update course in database;

return updated course;

return 404 if course does not exist.

Check:

STUDENT gets 403;

ADMIN updates course;

fake courseId gets 404;

invalid input gets 400.

Read:

PATCH semantics;

SQL UPDATE RETURNING;

route params.

### MG-43
Goal: add protected API for lesson create and lesson edit.

Routes:

POST /api/v1/admin/lessons

PATCH /api/v1/admin/lessons/{lessonId}

Rules:

only ADMIN;

STUDENT receives 403;

missing token receives 401;

title is required;

course_id must point to an existing course;

position must be greater than zero.

Check:

create lesson as ADMIN;

edit lesson as ADMIN;

try as STUDENT;

try invalid course_id;

try empty title.

### MG-46
Summary:
Add admin endpoint for groups

Epic:
MG-23

Description:
Goal:
Allow ADMIN users to create course groups through API.

Route:
POST /api/v1/admin/groups

Request fields:

course_id

title

start_date

end_date

Rules:

only ADMIN can call this endpoint;

STUDENT receives 403;

missing or invalid token returns 401;

course_id must point to an existing course;

title is required;

start_date and end_date are optional.

How to check:

Register user.

Make user ADMIN in local DB.

Login and get token.

Call POST /api/v1/admin/groups with valid course_id.

Expected result: group is created.

Call same endpoint as STUDENT.

Expected result: 403.

Call with fake course_id.

Expected result: 404.

Call with empty title.

Expected result: 400.

What to read:

SQL foreign keys

Go HTTP handlers

Role-based middleware

JSON request validation

### MG-47
Summary:
Document how to create first ADMIN user

Epic:
MG-23

Description:
Goal:
Explain how developers can create an ADMIN account for local development.

What to do:

add a section to README;

explain that public registration creates STUDENT users;

explain how to change a local user role to ADMIN through SQL;

explain that we do not expose a public create-admin API in MVP.

Example flow:

Register normal user through POST /api/v1/auth/register.

Open local database.

Change user role to ADMIN.

Login again.

Use returned token for admin endpoints.

How to check:

Register user.

Change role to ADMIN in DB.

Login.

Decode JWT.

Check that role is ADMIN.

Call an admin route.

Expected result: access allowed.

What to read:

PostgreSQL UPDATE statement

JWT claims

Role-based access control

README documentation examples

