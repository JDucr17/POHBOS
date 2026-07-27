-include .env
export

MIGRATIONS_DIR = infra/migrations
GO_MODULE      = github.com/JDucr17/streamline

PIPELINE_DIR        = services/pipeline
BIN_DIR             = bin

DATABASE_URL ?= postgres://streamline:streamline@localhost:5432/streamline?sslmode=disable
REDIS_URL    ?= redis://localhost:6379


PEEK_COUNT ?= 5

.PHONY: up down restart logs ps \
        airflow-up airflow-stop airflow-restart airflow-logs airflow-import-errors \
        migrate migrate-down migrate-status psql \
        kafka-topics list-topics peek-events peek-decisions peek-dlq \
        run-ingestor build-ingestor \
        run-event-sink run-decision-sink build-sink \
        run-detector build-detector \
        run-query-api build-queryapi \
        run-baseline-worker build-baseline-worker \
        run build clean

clean:
	rm -rf $(BIN_DIR)/*

up:
	docker compose up -d

down:
	docker compose down

restart: down up

logs:
	docker compose logs -f

ps:
	docker compose ps

migrate:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

psql:
	docker compose exec postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

kafka-topics:
	./infra/kafka/create_topics.sh

list-topics:
	docker compose exec redpanda rpk topic list

peek-events:
	docker compose exec redpanda rpk topic consume raw_events --num $(PEEK_COUNT)

peek-decisions:
	docker compose exec redpanda rpk topic consume decisions --num $(PEEK_COUNT)

peek-dlq:
	docker compose exec redpanda rpk topic consume dead_letter_events --num $(PEEK_COUNT)

run-ingestor:
	cd $(PIPELINE_DIR) && go run ./cmd/ingestor

build-ingestor:
	cd $(PIPELINE_DIR) && go build -o ../../$(BIN_DIR)/ingestor ./cmd/ingestor

run-event-sink:
	cd $(PIPELINE_DIR) && SINK_TARGET=events go run ./cmd/sink

run-decision-sink:
	cd $(PIPELINE_DIR) && SINK_TARGET=decisions go run ./cmd/sink

build-sink:
	cd $(PIPELINE_DIR) && go build -o ../../$(BIN_DIR)/sink ./cmd/sink

run-detector:
	cd $(PIPELINE_DIR) && go run ./cmd/detector

build-detector:
	cd $(PIPELINE_DIR) && go build -o ../../$(BIN_DIR)/detector ./cmd/detector

build-queryapi:
	cd $(PIPELINE_DIR) && go build -o ../../$(BIN_DIR)/queryapi ./cmd/queryapi

run-query-api:
	cd $(PIPELINE_DIR) && go run ./cmd/queryapi

run-baseline-worker:
	cd $(PIPELINE_DIR) && go run ./cmd/baseline-worker

build-baseline-worker:
	cd $(PIPELINE_DIR) && go build -o ../../$(BIN_DIR)/baseline-worker ./cmd/baseline-worker

build: build-ingestor build-sink build-detector build-queryapi build-baseline-worker

airflow-up:
	docker compose up -d --build airflow

airflow-stop:
	docker compose stop airflow

airflow-logs:
	docker compose logs -f airflow

airflow-import-errors:
	docker compose exec airflow airflow dags list-import-errors