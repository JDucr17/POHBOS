-include .env
export

MIGRATIONS_DIR = infra/migrations
DATABASE_URL ?= postgres://streamline:streamline@localhost:5432/streamline?sslmode=disable

.PHONY: up down migrate migrate-down migrate-status psql

up:
	docker compose up -d

down:
	docker compose down

migrate:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

psql:
	docker compose exec postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)