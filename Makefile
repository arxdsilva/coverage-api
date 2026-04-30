SHELL := /bin/sh

COMPOSE ?= docker compose
GOOSE ?= go run github.com/pressly/goose/v3/cmd/goose@latest
MIGRATIONS_DIR ?= ./migrations
SEEDS_FILE ?= ./seeds/001_seed.sql

DB_USER ?= coverage
DB_PASSWORD ?= coverage
DB_NAME ?= coverage
DB_PORT ?= 5432
DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: help deps fmt test run
.PHONY: compose-up compose-down compose-logs compose-ps
.PHONY: migrate-up migrate-down migrate-status migrate-reset migrate-create
.PHONY: seed seed-docker
.PHONY: coverage-file coverage-upload
.PHONY: frontend-run frontend-dev
.PHONY: update-branch

COVERAGE_PROFILE ?= coverage.out
COVERAGE_PAYLOAD ?= coverage-upload.json
API_URL ?= http://localhost:8080/v1/coverage-runs
API_KEY ?= dev-local-key
COVERAGE_PROJECT_KEY ?= github.com/arxdsilva/opencoverage
COVERAGE_PROJECT_NAME ?= coverage-api

help:
	@echo "Available targets:"
	@echo "  make deps               - Download dependencies"
	@echo "  make fmt                - Format Go files"
	@echo "  make test               - Run tests"
	@echo "  make run                - Run API locally"
	@echo "  make frontend-run       - Run frontend dashboard locally on :8090"
	@echo "  make frontend-dev       - Run API and frontend together (local)"
	@echo "  make compose-up         - Start db + migrate + seed + api + frontend via docker compose"
	@echo "                            Example with busy local 5432: DB_PORT=5433 make compose-up"
	@echo "                            If upgrading Postgres major versions: make compose-down then make compose-up"
	@echo "  make compose-down       - Stop docker compose services"
	@echo "  make compose-logs       - Tail compose logs"
	@echo "  make migrate-up         - Run DB migrations locally (requires local DB)"
	@echo "  make migrate-down       - Roll back one migration locally"
	@echo "  make migrate-status     - Show migration status locally"
	@echo "  make migrate-reset      - Roll back all migrations locally"
	@echo "  make migrate-create name=<migration_name> - Create new SQL migration"
	@echo "  make seed               - Run local seed SQL against local DB"
	@echo "  make seed-docker        - Run seed SQL through compose seed service"
	@echo "  make coverage-file      - Generate coverage.out and API payload JSON file"
	@echo "  make coverage-upload    - Generate + upload coverage payload to running API"
	@echo "  make update-branch      - Fetch and merge current branch with arxdsilva/main"

deps:
	go mod tidy

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

test:
	go test ./...

run:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/api

frontend-run:
	FRONTEND_ADDR=":8090" API_BASE_URL="http://localhost:8080" API_KEY_SECRET="$(API_KEY)" go run ./cmd/frontend

frontend-dev:
	@echo "Run in two terminals:"
	@echo "  1) DATABASE_URL=\"$(DATABASE_URL)\" API_KEY_SECRET=\"$(API_KEY)\" go run ./cmd/api"
	@echo "  2) FRONTEND_ADDR=\":8090\" API_BASE_URL=\"http://localhost:8080\" API_KEY_SECRET=\"$(API_KEY)\" go run ./cmd/frontend"

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

seed:
	psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f "$(SEEDS_FILE)"

seed-docker:
	$(COMPOSE) run --rm seed

coverage-file:
	go test ./... -coverprofile=$(COVERAGE_PROFILE)
	go run ./cmd/coveragecli \
		-coverprofile $(COVERAGE_PROFILE) \
		-out $(COVERAGE_PAYLOAD) \
		-project-key $(COVERAGE_PROJECT_KEY) \
		-project-name $(COVERAGE_PROJECT_NAME)

coverage-upload:
	go test ./... -coverprofile=$(COVERAGE_PROFILE)
	go run ./cmd/coveragecli \
		-coverprofile $(COVERAGE_PROFILE) \
		-out $(COVERAGE_PAYLOAD) \
		-project-key $(COVERAGE_PROJECT_KEY) \
		-project-name $(COVERAGE_PROJECT_NAME) \
		-api-url $(API_URL) \
		-api-key $(API_KEY) \
		-upload

update-branch:
	git fetch arxdsilva && git merge arxdsilva/main
