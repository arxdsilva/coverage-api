# coverage-api

Go REST API for ingesting coverage runs and computing coverage deltas.

## Architecture

This project follows Hexagonal Architecture (ports and adapters):

- `cmd/api` - application bootstrap
- `internal/domain` - entities and deterministic domain logic
- `internal/application` - use cases and ports
- `internal/adapters/http` - HTTP transport and middleware
- `internal/adapters/postgres` - repository implementations
- `internal/adapters/auth` - API key authentication adapter
- `internal/platform` - config and infrastructure utilities

## Requirements

- Go 1.23+
- PostgreSQL 14+

## Configuration

Environment variables:

- `DATABASE_URL` (required)
- `SERVER_ADDR` (default `:8080`)
- `API_KEY_HEADER` (default `X-API-Key`)
- `API_KEY_SECRET` (required; value expected in the API key header)
- `SHUTDOWN_TIMEOUT_SECONDS` (default `10`)

## Run

```bash
export DATABASE_URL="postgres://coverage:coverage@localhost:5432/coverage?sslmode=disable"
export API_KEY_SECRET="dev-local-key"
go run ./cmd/api
```

Start full local stack with Docker Compose (db + migrate + api):

```bash
make compose-up
```

If port `5432` is already in use on your machine, override it:

```bash
DB_PORT=5433 make compose-up
```

## Migrations

Initial schema is in `migrations/001_init.sql`.

Common migration commands:

```bash
make migrate-status
make migrate-up
make migrate-down
make migrate-create name=add_new_table
```

## API

Main endpoints:

- `POST /v1/coverage-runs`
- `GET /v1/projects/{projectId}`
- `GET /v1/projects/{projectId}/coverage-runs`
- `GET /v1/projects/{projectId}/coverage-runs/latest-comparison`

For full contract details, see `SPEC.md`.

## Usage (curl)

Set variables first:

```bash
export BASE_URL="http://localhost:8080"
export API_KEY="dev-local-key"
export PROJECT_ID="replace-with-project-id"
```

Health check (no auth):

```bash
curl -i "$BASE_URL/healthz"
```

Ingest a coverage run:

```bash
curl -i -X POST "$BASE_URL/v1/coverage-runs" \
	-H "Content-Type: application/json" \
	-H "X-API-Key: $API_KEY" \
	-d '{
		"projectKey": "org/repo-service",
		"projectName": "repo-service",
		"defaultBranch": "main",
		"branch": "main",
		"commitSha": "a1b2c3d4",
		"author": "alice",
		"triggerType": "push",
		"runTimestamp": "2026-03-28T12:00:00Z",
		"totalCoveragePercent": 83.42,
		"packages": [
			{"importPath": "github.com/acme/repo-service/internal/api", "coveragePercent": 85.10},
			{"importPath": "github.com/acme/repo-service/internal/service", "coveragePercent": 80.70}
		]
	}'
```

Get project metadata:

```bash
curl -i "$BASE_URL/v1/projects/$PROJECT_ID" \
	-H "X-API-Key: $API_KEY"
```

List coverage runs (paginated):

```bash
curl -i "$BASE_URL/v1/projects/$PROJECT_ID/coverage-runs?page=1&pageSize=20&branch=main" \
	-H "X-API-Key: $API_KEY"
```

Get latest comparison:

```bash
curl -i "$BASE_URL/v1/projects/$PROJECT_ID/coverage-runs/latest-comparison" \
	-H "X-API-Key: $API_KEY"
```
