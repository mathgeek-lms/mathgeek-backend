# Mathgeek Backend

## Seed Data

Run migrations first:

```sh
make migrate-up
```

Then load development seed data:

```sh
go run ./cmd/seed
```

The seed command uses `USERS_DB_DSN` from `.env`, creates or updates the `Algebra Basics` course, and creates or updates `Algebra Basics Group A` linked to that course.
