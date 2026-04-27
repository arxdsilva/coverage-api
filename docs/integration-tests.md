# Sending Integration Test Results to Coverage API

This guide explains how a CI job should run Ginkgo integration tests, produce a JSON report, and upload it to this API.

## What Happens on Upload

When CI sends a Ginkgo report to:

- `POST /v1/integration-test-runs`

the API will:

1. Resolve or auto-create the project by `projectKey`.
2. Persist the integration run and every spec result in one transaction.
3. Compute pass-rate comparison against the latest default-branch baseline.
4. Return a response with run summary and comparison.

Runs are persisted even when specs fail.

## Required Request Fields

- `projectKey`: stable repository identifier (for example: `owner/repo`)
- `branch`: branch name
- `commitSha`: commit SHA
- `triggerType`: `push`, `pr`, or `manual`
- `runTimestamp`: current UTC timestamp
- `environment` (optional): runtime environment (`test`, `stage`, or `prod`)
- `ginkgoReport`: normalized Ginkgo JSON payload (see below)

Authentication header:

- `X-API-Key: <your-api-secret>`

## Recommended CI Flow

### 1. Run Ginkgo with JSON output

```bash
ginkgo -r ./integration \
  --json-report=artifacts/ginkgo-integration-report.json \
  --junit-report=artifacts/ginkgo-integration-report.xml \
  --output-dir=artifacts
```

The `--json-report` flag is required. The CLI reads this file.

### 2. Upload with coveragecli

```bash
go run ./cmd/coveragecli integration-upload \
  -ginkgo-report artifacts/ginkgo-integration-report.json \
  -api-url "$INTEGRATION_API_URL" \
  -api-key "$COVERAGE_API_KEY" \
  -project-key "${GITHUB_REPOSITORY}" \
  -project-name "${GITHUB_REPOSITORY##*/}" \
  -project-group "team-name" \
  -default-branch "main" \
  -branch "${GITHUB_HEAD_REF:-${GITHUB_REF_NAME}}" \
  -commit-sha "${GITHUB_SHA}" \
  -author "github-actions" \
  -trigger-type push \
  -run-timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

The CLI validates the Ginkgo JSON before sending and exits non-zero on transport, auth, or validation failure.

## Response Fields to Use in CI Policy

- `run.status`: `passed` or `failed`
- `run.passRatePercent`: overall pass rate for this run
- `comparison.direction`: `up`, `down`, `equal`, or `new`
- `comparison.deltaPercent`: change in pass rate versus baseline (null on first run)
- `comparison.newFailures`: count of specs that newly failed in this run
- `comparison.resolvedFailures`: count of specs that were failing in baseline but passed here

## GitHub Actions Example

```yaml
- name: Run integration tests
  run: |
    ginkgo -r ./integration \
      --json-report=artifacts/ginkgo-integration-report.json \
      --output-dir=artifacts

- name: Upload integration results
  id: integration_upload
  if: ${{ secrets.INTEGRATION_API_URL != '' && secrets.COVERAGE_API_KEY != '' }}
  env:
    INTEGRATION_API_URL: ${{ secrets.INTEGRATION_API_URL }}
    COVERAGE_API_KEY: ${{ secrets.COVERAGE_API_KEY }}
  run: |
    go run ./cmd/coveragecli integration-upload \
      -ginkgo-report artifacts/ginkgo-integration-report.json \
      -api-url "$INTEGRATION_API_URL" \
      -api-key "$COVERAGE_API_KEY" \
      -project-key "${{ github.repository }}" \
      -project-name "${{ github.event.repository.name }}" \
      -project-group "your-team" \
      -default-branch "main" \
      -branch "${{ github.head_ref || github.ref_name }}" \
      -commit-sha "${{ github.sha }}" \
      -author "github-actions" \
      -trigger-type "${{ github.event_name == 'pull_request' && 'pr' || 'push' }}" \
      -run-timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
      > integration-api-response.json

    STATUS=$(jq -r '.run.status' integration-api-response.json)
    PASS_RATE=$(jq -r '.run.passRatePercent' integration-api-response.json)
    DIRECTION=$(jq -r '.comparison.direction' integration-api-response.json)
    DELTA=$(jq -r '.comparison.deltaPercent' integration-api-response.json)
    NEW_FAILURES=$(jq -r '.comparison.newFailures' integration-api-response.json)

    echo "status=$STATUS" >> "$GITHUB_OUTPUT"
    echo "passRate=$PASS_RATE" >> "$GITHUB_OUTPUT"
    echo "direction=$DIRECTION" >> "$GITHUB_OUTPUT"
    echo "delta=$DELTA" >> "$GITHUB_OUTPUT"
    echo "newFailures=$NEW_FAILURES" >> "$GITHUB_OUTPUT"

- name: Warn on integration failures
  if: ${{ steps.integration_upload.outputs.status == 'failed' }}
  run: |
    echo "::warning title=Integration Tests Failed::PassRate=${{ steps.integration_upload.outputs.passRate }} Delta=${{ steps.integration_upload.outputs.delta }} NewFailures=${{ steps.integration_upload.outputs.newFailures }}"

- name: Fail job when integration tests fail
  if: ${{ steps.integration_upload.outputs.status == 'failed' }}
  run: |
    echo "Integration tests failed. See failed specs in the dashboard."
    exit 1
```

Remove the final `Fail job` step if you want CI to warn without blocking the merge.

## Uploading Without the CLI (Manual curl)

If you cannot use `coveragecli`, construct the JSON payload manually and POST it directly.

Minimal payload shape:

```json
{
  "projectKey": "owner/repo",
  "projectName": "repo",
  "projectGroup": "team-name",
  "defaultBranch": "main",
  "branch": "main",
  "commitSha": "a1b2c3d4",
  "author": "github-actions",
  "triggerType": "push",
  "runTimestamp": "2026-04-25T12:00:00Z",
  "ginkgoReport": {
    "suiteDescription": "integration suite",
    "suitePath": "./integration",
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

```bash
curl -sS -X POST "$INTEGRATION_API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $COVERAGE_API_KEY" \
  --data-binary @integration-payload.json > integration-api-response.json
```

Accepted `state` values: `passed`, `failed`, `skipped`, `pending`, `panicked`, `interrupted`, `timedout`, `flaked`.

## Quick Answers

- Will runs be stored even when specs fail? Yes.
- Can CI warn instead of fail? Yes — omit the exit 1 step.
- Is a project auto-created if it does not exist? Yes.
- Is `failure` required when a spec fails? Yes — the API rejects failed specs without a failure object.
- What baseline is used for comparison? The latest run on the project's default branch.
- What happens on the first run? `direction` is `new` and `deltaPercent` is null.
