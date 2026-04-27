# Integration Tests Specification (API + Frontend)

## 1. Overview

### 1.1 Purpose
Define how coverage-api accepts, stores, and presents integration test results in addition to existing unit test and coverage workflows.

### 1.2 Primary Outcome
Provide a deterministic way to answer:
1. Did integration tests pass for this run?
2. Which specs failed, skipped, or flaked?
3. How does the latest run compare with the previous baseline run on the default branch?
4. How should CI and the frontend consume the same source of truth?
5. How can teams upload integration test data with the same `coveragecli` used for coverage uploads?

### 1.3 Source of Truth
The canonical input format is Ginkgo output, specifically a Ginkgo v2 JSON report artifact.

## 2. Scope

### 2.1 In Scope (v1)
1. Ingest integration test run metadata and summarized suite/spec results.
2. Parse and normalize Ginkgo JSON report fields at ingest boundary.
3. Compute latest-vs-baseline comparison on default branch.
4. Expose API endpoints for run history and latest comparison.
5. Render integration test status in frontend dashboard.
6. Keep API key auth on all API endpoints.

### 2.2 Out of Scope (v1)
1. Storing full raw stdout for every spec node.
2. Video/screenshots/artifacts from browser test runners.
3. Flake auto-retry orchestration by the API.
4. Cross-project aggregated analytics beyond basic totals.

## 3. Key Decisions and Assumptions

1. Existing coverage ingest remains unchanged.
2. Integration tests are represented as a separate run type and data model.
3. Baseline is latest run on project default branch.
4. Missing baseline still returns HTTP 200 with comparison direction `new`.
5. CI uploads one Ginkgo JSON report per integration run.
6. API supports idempotency only by storing all runs (no dedupe), same as coverage runs.

## 4. Data Model (Draft)

### 4.1 Entity Summary
1. IntegrationTestRun
2. IntegrationSpecResult

### 4.2 Logical Fields

#### IntegrationTestRun
1. id (UUID)
2. project_id (UUID, FK -> projects.id)
3. branch (string)
4. commit_sha (string)
5. author (string, optional)
6. trigger_type (enum/string: push, pr, manual)
7. run_timestamp (timestamp)
8. ginkgo_version (string, optional)
9. suite_description (string)
10. suite_path (string)
11. total_specs (integer)
12. passed_specs (integer)
13. failed_specs (integer)
14. skipped_specs (integer)
15. flaked_specs (integer)
16. pending_specs (integer)
17. interrupted (boolean)
18. timed_out (boolean)
19. duration_ms (integer)
20. environment (enum/string, optional: test, stage, prod) - deployment environment
21. created_at (timestamp)

#### IntegrationSpecResult
1. id (UUID)
2. integration_run_id (UUID, FK -> integration_test_runs.id)
3. spec_path (text, normalized full container/spec text)
4. leaf_node_text (text)
5. state (enum/string: passed, failed, skipped, pending, flaky)
6. duration_ms (integer)
7. failure_message (text, nullable)
8. failure_location_file (text, nullable)
9. failure_location_line (integer, nullable)

## 5. API Contract (v1)

### 5.1 Authentication
1. Header: `X-API-Key: <key>`
2. Required for all endpoints.

### 5.2 Content Type
1. Request: `application/json`
2. Response: `application/json`

### 5.3 Endpoint Summary
1. `POST /v1/integration-test-runs` - ingest Ginkgo run and return comparison.
2. `GET /v1/projects/{projectId}/integration-test-runs` - paginated history.
3. `GET /v1/projects/{projectId}/integration-test-runs/latest-comparison` - latest run comparison.
4. `GET /v1/projects/{projectId}/integration-test-runs/{runId}` - full run details with failed specs.

## 6. POST /v1/integration-test-runs

### 6.1 Request Body (Normalized)
```json
{
  "projectKey": "org/repo-service",
  "projectName": "repo-service",
  "projectGroup": "platform-team",
  "defaultBranch": "main",
  "branch": "main",
  "commitSha": "a1b2c3d4",
  "author": "alice",
  "triggerType": "push",
  "runTimestamp": "2026-04-25T12:00:00Z",
  "environment": "prod",
  "ginkgoReport": {
    "suiteDescription": "integration suite",
    "suitePath": "./integration",
    "suiteConfig": {
      "randomSeed": 1714046400,
      "randomizeAllSpecs": false,
      "focusStrings": [],
      "skipStrings": []
    },
    "specReports": [
      {
        "leafNodeText": "creates a project on first ingest",
        "containerHierarchyTexts": ["POST /v1/coverage-runs"],
        "state": "passed",
        "runTime": 0.112
      },
      {
        "leafNodeText": "rejects invalid api key",
        "containerHierarchyTexts": ["auth middleware"],
        "state": "failed",
        "runTime": 0.053,
        "failure": {
          "message": "expected 401, got 500",
          "location": {
            "fileName": "integration/auth_test.go",
            "lineNumber": 84
          }
        }
      }
    ]
  }
}
```

### 6.2 Accepted Ginkgo Field Mapping
1. `suiteDescription` <- report `SuiteDescription`
2. `suitePath` <- report `SuitePath`
3. `specReports[*].leafNodeText` <- `LeafNodeText`
4. `specReports[*].containerHierarchyTexts` <- `ContainerHierarchyTexts`
5. `specReports[*].state` <- `State`
6. `specReports[*].runTime` <- `RunTime`
7. `specReports[*].failure.*` <- `Failure.*` when present

State normalization rules:
1. `passed` -> passed
2. `failed` -> failed
3. `skipped` -> skipped
4. `pending` -> pending
5. `panicked|interrupted|timedout` -> failed
6. `flaked` (from additional attempt metadata) -> flaky

### 6.3 Behavior
1. Resolve or auto-create project by `projectKey`.
2. Validate and normalize Ginkgo JSON payload.
3. Persist integration run and per-spec results in one DB transaction.
4. Lookup baseline from latest default branch integration run.
5. Compute deltas on totals and pass rate.
6. Return HTTP 200 with run summary and comparison.

### 6.4 Response Body (Example)
```json
{
  "project": {
    "id": "5d6e8f6d-f1c8-4f3f-8c93-caf78e7a6a34",
    "projectKey": "org/repo-service",
    "name": "repo-service",
    "defaultBranch": "main",
    "created": false
  },
  "run": {
    "id": "1afef22f-2f8f-42de-9f48-f2a11f32a044",
    "branch": "main",
    "commitSha": "a1b2c3d4",
    "runTimestamp": "2026-04-25T12:00:00Z",
    "totalSpecs": 120,
    "passedSpecs": 116,
    "failedSpecs": 2,
    "skippedSpecs": 2,
    "flakedSpecs": 1,
    "passRatePercent": 96.67,
    "durationMs": 18203,
    "status": "failed"
  },
  "comparison": {
    "baselineSource": "latest_default_branch",
    "previousPassRatePercent": 98.10,
    "currentPassRatePercent": 96.67,
    "deltaPercent": -1.43,
    "direction": "down",
    "newFailures": 2,
    "resolvedFailures": 1
  },
  "failedSpecs": [
    {
      "specPath": "auth middleware > rejects invalid api key",
      "failureMessage": "expected 401, got 500",
      "file": "integration/auth_test.go",
      "line": 84
    }
  ]
}
```

First-run case:
1. `previousPassRatePercent = null`
2. `deltaPercent = null`
3. `direction = "new"`

## 7. History and Comparison Endpoints

### 7.1 GET /v1/projects/{projectId}/integration-test-runs

Query params:
1. `page` (default 1)
2. `pageSize` (default 20, max 100)
3. `branch` (optional)
4. `from` (optional ISO timestamp)
5. `to` (optional ISO timestamp)
6. `status` (optional `passed|failed`)

Response:
```json
{
  "items": [
    {
      "id": "1afef22f-2f8f-42de-9f48-f2a11f32a044",
      "branch": "main",
      "commitSha": "a1b2c3d4",
      "runTimestamp": "2026-04-25T12:00:00Z",
      "totalSpecs": 120,
      "failedSpecs": 2,
      "passRatePercent": 96.67,
      "status": "failed"
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "totalItems": 43,
    "totalPages": 3
  }
}
```

### 7.2 GET /v1/projects/{projectId}/integration-test-runs/latest-comparison
Returns latest run summary and latest-vs-baseline comparison.

### 7.3 GET /v1/projects/{projectId}/integration-test-runs/{runId}
Returns one run with full failed spec list and summary metrics.

### 7.4 GET /v1/integration-test-runs/heatmap
Returns recent integration runs grouped by project, intended for the all-project heatmap view.

Query params:
1. `runsPerProject` (optional, default 10, max 30) — number of most recent runs to return per project.
2. `branch` (optional) — filter runs to a specific branch name across all projects.
3. `status` (optional `passed|failed`) — filter runs by status across all projects.

Response:
```json
{
  "groups": [
    {
      "groupName": "platform-team",
      "projects": [
        {
          "projectId": "5d6e8f6d-f1c8-4f3f-8c93-caf78e7a6a34",
          "projectName": "repo-service",
          "projectKey": "org/repo-service",
          "runs": [
            {
              "id": "1afef22f-2f8f-42de-9f48-f2a11f32a044",
              "branch": "main",
              "commitSha": "a1b2c3d4",
              "runTimestamp": "2026-04-25T12:00:00Z",
              "passRatePercent": 96.67,
              "status": "failed"
            }
          ]
        }
      ]
    },
    {
      "groupName": "",
      "projects": [
        {
          "projectId": "9a1b2c3d-0000-4000-8000-000000000001",
          "projectName": "ungrouped-service",
          "projectKey": "org/ungrouped-service",
          "runs": []
        }
      ]
    }
  ]
}
```

Behavior:
1. Returns all projects that have at least one integration run, grouped by their `group` field.
2. Groups are ordered alphabetically by `groupName`; projects within a group are ordered alphabetically by `projectName`.
3. Projects with no `group` value are collected into a group with `groupName: ""` placed last.
4. Runs within each project are ordered newest first.
5. Projects with no matching runs after filter application are omitted from the response (and their group is omitted if it becomes empty).
6. Requires API key auth.

## 8. Frontend Specification

### 8.0 Navigation Model
1. Integration test UI must be a dedicated screen (separate route), not an inline section on the homepage.
2. Main page must expose a clear entry point to this screen (for example: `Integration Tests` button/card/link in the primary dashboard navigation).
3. Integration screen must include a persistent `Back to Home` action that returns to the main dashboard route.
4. Browser back/forward navigation must work correctly between home and integration routes.

### 8.1 Dashboard Additions
1. Home dashboard shows a compact `Integration Test Health` summary card.
2. Card includes latest run status (`passed`/`failed`), run success ratio percentage (`passed runs / failed runs * 100`), failed count, and duration.
3. Card includes delta badge versus baseline (`up`, `down`, `equal`, `new`).
4. Card includes `Open Integration Screen` action that navigates to the dedicated integration route.

### 8.2 New Views
1. Dedicated integration screen contains `Integration Runs` table for selected project.
2. Dedicated integration screen includes `Failed Specs` drawer/modal for selected run.
3. Dedicated integration screen includes filters: branch, date range, and status.
4. Dedicated integration screen header includes `Back to Home` control.
5. Dedicated integration screen includes a horizontal run-chain graphic where each node is a run, ordered oldest-to-newest with newest on the right, colored green for `passed` and red for `failed`.
6. Clicking a run-chain node selects that run and updates the failed-spec details pane.
7. Run-chain display is capped to at most 5 runs (the newest 5 from the active run-list query).
8. Integration pass-rate summary card reflects run success ratio percentage (`passedRuns / failedRuns * 100`) over the returned run-list window (up to 20 runs), not only the latest single run.

### 8.3 Integration Heatmap (All-Project View)
1. Dedicated integration screen includes an `Integration Heatmap` region that visualizes recent integration runs for **all projects simultaneously**, organized by project group.
2. Heatmap layout: projects are grouped by their `group` field. Each group is rendered as a labeled section. Within each group, one row per project; one tile per run. Runs are ordered oldest-to-newest within each project row (newest on the right).
3. Groups are ordered alphabetically. Projects within a group are ordered alphabetically. Projects with no group appear in an unlabeled section at the bottom.
4. Heatmap data source is `GET /api/integration-test-runs/heatmap` — the dedicated all-project aggregated endpoint (see §7.4).
5. Heatmap is rendered using each project's default branch only.
6. The heatmap status filter is independent from per-project run-table filters. The `runsPerProject` count is controlled by the frontend and optimized for visualization density (default 10).
7. Tile status semantics use emoji markers:
  - `✅` for `passed`
  - `❌` for `failed`
8. Heatmap project-row background tint reflects the newest run result for that project.
9. Hover/focus state must show at minimum: project name, group, run id, branch, commit sha, timestamp, status, pass rate.
10. Clicking a heatmap tile for the currently selected project must synchronize selected run state with the run chain, run table, and failed-spec details. Clicking a tile for a different project may switch the active project selection.
11. Heatmap supports local reload and its own status filter, independent from the per-project run table.

### 8.3.1 Integration Refresh Semantics
1. The integration route must expose a primary screen-level `Refresh` action.
2. A screen-level refresh must refetch the project catalog and then refresh the active project's latest comparison, run list, run chain, and failed-spec details.
3. If the active project still exists after refresh, it must remain selected.
4. If the selected run still exists after refresh, it should remain selected; otherwise the newest run in the refreshed list becomes selected.
5. The per-project `Reload` control must refresh only the active project's comparison, run list, run chain, and failed-spec details.
6. The integration heatmap overlay `Reload` control must refresh only the all-project heatmap query.
7. Heatmap overlay status filter must be preserved across both the overlay-local reload and the screen-level refresh.
8. If the heatmap overlay is open during a screen-level refresh, it should remain open and repaint when the refreshed heatmap data arrives.
9. Refresh operations must not reset route-level query parameter behavior such as `?heatmap=open`.
10. Error handling must remain local to the failing region whenever possible; for example, a heatmap reload failure must not clear the per-project run table.
11. The integration route must expose an auto-refresh configuration control with these intervals: `off`, `15s`, `30s`, `60s`, `5m`.
12. When auto refresh is enabled, each interval tick must execute the same screen-level refresh workflow as the primary `Refresh` action.
13. The selected auto-refresh interval should persist for the integration route independently of the home route.
14. Auto refresh must preserve active project selection, selected run selection when still present, heatmap open state, heatmap filters, and route query parameter behavior.
15. If an automatic refresh tick occurs while another integration refresh is already in progress, the overlapping tick must be skipped.
16. Auto refresh should pause when the browser tab is hidden and resume when the tab becomes visible again.
17. Automatic refresh failures must not disable future scheduled refresh attempts unless the user turns auto refresh off.
18. The integration route must display a visible auto-refresh status line near the interval control.
19. When auto refresh is enabled, the status line must display a live countdown to the next automatic refresh.
20. The status line must reflect runtime states explicitly:
  - off
  - refreshing now
  - paused while tab is hidden
21. Status/countdown updates must occur continuously without requiring additional user input.

### 8.4 Frontend Proxy Routes
The following proxy routes are required:
1. `GET /api/projects/{projectId}/integration-test-runs`
2. `GET /api/projects/{projectId}/integration-test-runs/latest-comparison`
3. `GET /api/projects/{projectId}/integration-test-runs/{runId}`
4. `GET /api/integration-test-runs/heatmap` — new route required for the all-project heatmap; must be proxied to `GET /v1/integration-test-runs/heatmap` with API key injected.

## 9. Ginkgo Execution and Artifact Contract

### 9.1 CI Command (Recommended)
```bash
ginkgo -r ./integration \
  --json-report=artifacts/ginkgo-integration-report.json \
  --junit-report=artifacts/ginkgo-integration-report.xml \
  --output-dir=artifacts
```

### 9.2 Upload Contract
The extended `coveragecli` command must:
1. Read `ginkgo-integration-report.json`.
2. Construct `POST /v1/integration-test-runs` payload.
3. Include metadata: `projectKey`, `branch`, `commitSha`, `triggerType`, `runTimestamp`.
4. Fail pipeline only if upload call fails or if policy requires failing on integration status.

### 9.3 coveragecli Extension (Required)
The repository CLI will be extended (not replaced) to support integration test upload.

Proposed command shape:
```bash
go run ./cmd/coveragecli integration-upload \
  -ginkgo-report artifacts/ginkgo-integration-report.json \
  -api-url http://localhost:8080/v1/integration-test-runs \
  -api-key dev-local-key \
  -project-key org/repo-service \
  -project-name repo-service \
  -project-group platform-team \
  -default-branch main \
  -branch main \
  -commit-sha a1b2c3d4 \
  -author alice \
  -trigger-type push \
  -run-timestamp 2026-04-25T12:00:00Z
```

CLI requirements:
1. Reuse existing API auth/header behavior from coverage upload flow.
2. Reuse project metadata flags where possible (`project-key`, `project-name`, `project-group`, `default-branch`).
3. Validate ginkgo report readability and JSON shape before sending request.
4. Return non-zero exit code for transport/auth/validation failures.
5. Optionally print response summary (`status`, `passRatePercent`, `deltaPercent`) for CI logs.

## 10. Validation Rules

1. `projectKey`, `branch`, `commitSha`, `triggerType`, `runTimestamp` required.
2. `ginkgoReport.specReports` required and non-empty.
3. `state` must be one of accepted Ginkgo states.
4. `runTime` must be >= 0.
5. `failure` object required when normalized state is `failed`.

## 11. Comparison Rules

1. Baseline is latest integration run on default branch.
2. `passRatePercent = (passed_specs / total_specs) * 100` rounded to 2 decimals.
3. `deltaPercent = currentPassRatePercent - previousPassRatePercent`.
4. Direction:
   - `up` if delta > 0
   - `down` if delta < 0
   - `equal` if delta = 0
   - `new` if no baseline exists
5. Run status:
   - `passed` if `failed_specs == 0` and `interrupted == false` and `timed_out == false`
   - `failed` otherwise

## 12. Error Model

Structured error payload stays consistent with coverage API:
```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "ginkgoReport.specReports must not be empty",
    "details": {
      "field": "ginkgoReport.specReports"
    }
  }
}
```

Standard codes:
1. `UNAUTHENTICATED` (401)
2. `INVALID_ARGUMENT` (400)
3. `NOT_FOUND` (404)
4. `INTERNAL` (500)

## 13. Testing Strategy for This Feature

1. Domain tests:
   - state normalization
   - pass-rate and delta math
   - first-run behavior
2. Use-case tests with mocked ports:
   - ingest run success
   - invalid payload mapping
   - baseline comparison logic
3. PostgreSQL adapter integration tests:
   - transactional insert of run + spec results
   - history query with pagination/filters
4. HTTP handler tests:
   - validation and error mapping
   - response shape consistency
5. Frontend tests:
   - proxy route coverage
   - rendering of status badge and failed specs table

## 14. Acceptance Criteria

1. API ingests Ginkgo JSON report for integration runs.
2. New projects are auto-created (same behavior as coverage ingest).
3. Integration run + spec results are persisted transactionally.
4. Latest response includes pass-rate comparison against default-branch baseline.
5. Frontend shows integration status and run history for selected project.
6. Frontend can display failed specs with file and line when available.
7. All new endpoints require API key auth at API boundary.
8. Existing coverage endpoints remain backward compatible.
9. `coveragecli` supports integration test report upload to `POST /v1/integration-test-runs`.
10. Frontend exposes a dedicated integration tests route reachable from the main page.
11. Users can navigate from integration screen back to the homepage via explicit UI control.
12. Integration screen includes an all-project heatmap powered by `GET /v1/integration-test-runs/heatmap`, displaying one row per project with pass/fail color semantics and synchronized selection behavior with run details for the active project.

## 15. Rollout Notes

1. Ship API endpoints behind feature flag if needed (`ENABLE_INTEGRATION_TESTS=true`).
2. Land DB migration first, then API adapters, then frontend UI.
3. Extend `coveragecli`, then add CI job to publish Ginkgo JSON artifact and upload via CLI.
4. Measure ingest latency and DB row growth before enabling by default.
