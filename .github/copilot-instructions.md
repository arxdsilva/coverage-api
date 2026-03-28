# Copilot Instructions for coverage-api

## Mission
Build and evolve a Go REST API for coverage ingestion and comparison using Hexagonal Architecture, clean boundaries, and production-grade engineering practices.

## Core Architecture: Hexagonal (Ports and Adapters)
Always organize code by domain boundaries and dependency direction.

1. Domain is at the center and knows nothing about frameworks, transport, or databases.
2. Application layer orchestrates use cases and depends only on domain + ports.
3. Adapters implement ports for HTTP, PostgreSQL, auth, and observability.
4. Dependencies point inward only.

### Required Layers
Use this structure (or equivalent naming):

- `cmd/api` - entrypoint and bootstrap wiring
- `internal/domain` - entities, value objects, domain services, domain errors
- `internal/application` - use cases and port interfaces
- `internal/adapters/http` - handlers, request/response DTOs, middleware
- `internal/adapters/postgres` - repository implementations
- `internal/adapters/auth` - API key validation adapter
- `internal/platform` - config, logger, metrics, tracing, db client, clock/uuid abstractions

Do not place business rules in handlers, SQL repositories, or middleware.

## Go Language Standards

1. Target the current stable Go version used by the project.
2. Keep packages cohesive and small.
3. Prefer composition over inheritance-like patterns.
4. Return errors, do not panic in normal control flow.
5. Wrap errors with context using `fmt.Errorf("...: %w", err)`.
6. Keep interfaces near the consumer (application ports), not global interface files.
7. Accept `context.Context` as first parameter for request-scoped operations.
8. Avoid package-level mutable state.
9. Keep functions short and intention-revealing.
10. Use constructor functions with explicit dependencies.

## Domain and Use Case Rules

1. Model explicit domain concepts:
   - Project
   - CoverageRun
   - PackageCoverage
   - Threshold evaluation and delta calculation
2. Domain types should enforce invariants (e.g., coverage must be between 0 and 100).
3. Use use-case services for workflows (ingest run, compute comparison, list history).
4. Use deterministic logic for comparison:
   - baseline = latest run on default branch
   - compute overall and package deltas
   - assign direction (`up`, `down`, `equal`, `new`)
5. Return typed application errors that adapters map to HTTP status codes.

## Ports (Application Interfaces)
Define ports in the application layer, for example:

- `ProjectRepository`
- `CoverageRunRepository`
- `PackageCoverageRepository`
- `APIKeyAuthenticator`
- `Clock`
- `IDGenerator`
- `TransactionManager`

Adapters implement these ports. Use cases must not import adapter packages.

## HTTP/API Conventions

1. Follow REST with `/v1` prefix.
2. Use JSON DTOs at the HTTP boundary only.
3. Validate input at boundary; enforce domain invariants inside domain.
4. Always return structured error payloads:
   - `{"error": {"code": "...", "message": "...", "details": {...}}}`
5. Return consistent response shapes and field naming.
6. Include request ID/correlation ID in logs and optionally response headers.

## Persistence (PostgreSQL) Practices

1. Use repository adapters for all DB access.
2. Use transactions for multi-step writes (run + packages).
3. Keep SQL explicit and indexed for key read paths.
4. Do not leak SQL row models into domain/application layers.
5. Keep migrations versioned and reversible where practical.
6. Handle `NULL` safely and map to domain optional values deliberately.

## Configuration and Bootstrapping

1. Use environment-driven config with explicit struct validation at startup.
2. Fail fast on invalid config.
3. Centralize wiring in `cmd/api`.
4. Inject dependencies; avoid service locators.

## Observability and Operations

1. Structured logging (JSON in production).
2. Log at appropriate levels; never log secrets or raw API keys.
3. Expose basic metrics (request count, latency, error rate).
4. Add readiness/liveness endpoints when infrastructure needs them.

## Security Requirements

1. Enforce API key auth on protected endpoints.
2. Load API key secret from environment (e.g., `API_KEY_SECRET`) and never log it.
3. Compare API keys using constant-time operations where applicable.
4. Sanitize and validate all external input.
5. Avoid verbose internal error leakage in HTTP responses.

## Testing Strategy

1. Prefer table-driven tests.
2. Unit test domain logic thoroughly (delta math, threshold status, edge cases).
3. Unit test use cases with mocked ports.
4. Add integration tests for PostgreSQL adapters.
5. Add HTTP handler tests for validation and error mapping.
6. Keep tests deterministic (inject clock and ID generator).

## Performance and Reliability

1. Keep handlers non-blocking and context-aware.
2. Add sensible timeouts for DB and HTTP server.
3. Avoid N+1 query patterns in history/list endpoints.
4. Design pagination with stable ordering.

## Code Generation and Edits
When generating or editing code in this repository:

1. Preserve hexagonal boundaries.
2. Add or update tests for behavior changes.
3. Keep public API behavior aligned with `SPEC.md`.
4. Do not introduce framework lock-in into domain/application layers.
5. Prefer small, reviewable commits and focused changes.

## Definition of Done for Changes
A change is complete only when:

1. Code follows architecture boundaries.
2. Error handling and input validation are explicit.
3. Tests cover happy paths and key edge/error cases.
4. Logging/metrics impact is considered.
5. API and schema changes are reflected in docs/migrations as needed.
