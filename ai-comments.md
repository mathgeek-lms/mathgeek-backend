### MG-44
Goal: make all API errors look the same.

Current problem:
Handlers and middleware may return errors in different JSON shapes.

Do:

create shared response helpers;

use one JSON format for errors;

update auth handlers;

update JWT middleware;

update role middleware;

add basic tests.

Format:
error.code and error.message.

Check:

validation error has same shape;

unauthorized has same shape;

forbidden has same shape;

not found has same shape.

Read:

HTTP status codes;

JSON API design;

Go helper packages.



### MG-45
Goal: cover the main backend flow with integration tests.

Do:

test auth endpoints;

test current user endpoint;

test course and lesson read endpoints;

test enrollment endpoint;

test lesson progress endpoints;

clean test data between runs.

Check:

tests pass locally;

tests pass in CI;

failed API behavior is covered too.

Read:

Go httptest;

testcontainers-go;

table-driven tests.

