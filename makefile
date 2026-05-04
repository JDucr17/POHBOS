-include .env
export

# Service abbreviations:
#   ep   = event-processor

MIGRATIONS_DIR      = infra/migrations
PROTO_DIR           = schemas
GO_PROTO_OUT        = services/event-processor/internal/extractor/proto
GO_MODULE           = github.com/JDucr17/streamline
EP_DIR              = services/event-processor
QAPI_DIR            = services/query-api
BIN_DIR             = bin

DATABASE_URL ?= postgres://streamline:streamline@localhost:5432/streamline?sslmode=disable

.PHONY: up down migrate migrate-down migrate-status psql proto \
        run-ep build-ep

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

proto:
	mkdir -p $(GO_PROTO_OUT)
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=. \
		--go_opt=module=$(GO_MODULE) \
		$(PROTO_DIR)/streamline/v1/feature_registry.proto

run-ep:
	cd $(EP_DIR) && go run ./cmd

build-ep:
	cd $(EP_DIR) && go build -o $(BIN_DIR)/event-processor ./cmd

build: build-ep build-qapi