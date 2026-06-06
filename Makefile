include .env
export

migrate-up:
	@echo "==> Running migrations up"
	@which goose > /dev/null || (echo "goose not found")
	goose -dir ./migrations postgres "$(USERS_DB_DSN)" up

migrate-down:
	@echo "==> Rolling back last migration"
	goose -dir ./migrations postgres "$(USERS_DB_DSN)" down

migrate-create:
ifndef NAME
	$(error NAME is required. Usage: make migrate-create NAME=migration-name)
endif
	@echo "==> Creating migration"
	goose -dir ./migrations create $(NAME) sql
	@echo "==> Created in ./migrations"