SHELL := /bin/sh

COMPOSE ?= docker compose
GOOSE ?= go run github.com/pressly/goose/v3/cmd/goose@latest
MIGRATIONS_DIR ?= ./migrations

DB_USER ?= coverage
DB_PASSWORD ?= coverage
DB_NAME ?= coverage
DB_PORT ?= 5432
DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable
COMPOSE_DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@db:5432/$(DB_NAME)?sslmode=disable

.PHONY: help deps fmt test run
.PHONY: compose-up compose-down compose-logs compose-ps
.PHONY: migrate-up migrate-down migrate-status migrate-reset migrate-create
.PHONY: migrate-up-docker migrate-down-docker

help:
	@echo "Available targets:"
	@echo "  make deps               - Download dependencies"
	@echo "  make fmt                - Format Go files"
	@echo "  make test               - Run tests"
	@echo "  make run                - Run API locally"
	@echo "  make compose-up         - Start db + migrate + api via docker compose"
	@echo "                            Example with busy local 5432: DB_PORT=5433 make compose-up"
	@echo "  make compose-down       - Stop docker compose services"
	@echo "  make compose-logs       - Tail compose logs"
	@echo "  make migrate-up         - Run DB migrations locally (requires local DB)"
	@echo "  make migrate-down       - Roll back one migration locally"
	@echo "  make migrate-status     - Show migration status locally"
	@echo "  make migrate-reset      - Roll back all migrations locally"
	@echo "  make migrate-create name=<migration_name> - Create new SQL migration"
	@echo "  make migrate-up-docker  - Run migrations against compose DB"
	@echo "  make migrate-down-docker - Roll back one migration against compose DB"

deps:
	go mod tidy

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

test:
	go test ./...

run:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/api

compose-up:
	$(COMPOSE) up -d --build

compose-down:
	$(COMPOSE) down -v

compose-logs:
	$(COMPOSE) logs -f

compose-ps:
	$(COMPOSE) ps

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

migrate-reset:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" reset

migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=<migration_name>"; exit 1; fi
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(name) sql

migrate-up-docker:
	$(COMPOSE) run --rm migrate up

migrate-down-docker:
	$(COMPOSE) run --rm migrate down
