# GitHub Actions Integration

This guide shows how to configure GitHub Actions to run Go unit tests, generate a coverage report, and upload it to a self-hosted coverage-api instance.

## Prerequisites

- Self-hosted coverage-api instance running (see [SELF_HOSTING.md](SELF_HOSTING.md))
- GitHub repository with Go code
- Repository secrets configured

## Setup

### 1. Configure Repository Secrets

In your GitHub repository settings, add:

1. **Settings > Secrets and variables > Actions > New repository secret**

- `COVERAGE_API_URL`: Full endpoint URL
  ```
  http://your-server:8080/v1/coverage-runs
  ```
  (For private networks, use internal IP or hostname)

- `COVERAGE_API_KEY`: The API key
  ```
  (Same as API_KEY_SECRET from coverage-api deployment)
  ```

### 2. Configure Optional Repository Variables

If you want to avoid hardcoding project metadata in the workflow, add GitHub Actions repository variables:

- `COVERAGE_PROJECT_KEY` (optional)
- `COVERAGE_PROJECT_NAME` (optional)
- `COVERAGE_PROJECT_GROUP` (optional)

These can be set in:

1. **Settings > Secrets and variables > Actions > Variables**

If omitted, the examples below fall back to repository-derived defaults.

## Workflow Example

Create `.github/workflows/coverage.yml`:

```yaml
name: Unit Tests & Coverage

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  COVERAGE_PROJECT_KEY: ${{ vars.COVERAGE_PROJECT_KEY }}
  COVERAGE_PROJECT_NAME: ${{ vars.COVERAGE_PROJECT_NAME }}
  COVERAGE_PROJECT_GROUP: ${{ vars.COVERAGE_PROJECT_GROUP }}

jobs:
  test:
    name: Test & Upload Coverage
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'

      - name: Run tests with coverage
        run: go test ./... -coverprofile=coverage.out

      - name: Install coverage CLI
        run: go install github.com/arxdsilva/opencoverage/cmd/coveragecli@latest

      - name: Generate coverage payload
        run: |
          PROJECT_KEY="${COVERAGE_PROJECT_KEY:-${{ github.repository }}}"
          PROJECT_NAME="${COVERAGE_PROJECT_NAME:-${{ github.event.repository.name }}}"
          PROJECT_GROUP="${COVERAGE_PROJECT_GROUP:-backend}"

          coveragecli \
            -coverprofile coverage.out \
            -out coverage-upload.json \
            -project-key "$PROJECT_KEY" \
            -project-name "$PROJECT_NAME" \
            -project-group "$PROJECT_GROUP" \
            -branch "${{ github.ref_name }}" \
            -commit-sha "${{ github.sha }}" \
            -author "${{ github.actor }}" \
            -trigger-type "push"

      - name: Upload to coverage-api
        if: ${{ secrets.COVERAGE_API_URL != '' && secrets.COVERAGE_API_KEY != '' }}
        run: |
          curl -X POST "${{ secrets.COVERAGE_API_URL }}" \
            -H "Content-Type: application/json" \
            -H "X-API-Key: ${{ secrets.COVERAGE_API_KEY }}" \
            --data-binary @coverage-upload.json

      - name: Archive coverage report
        uses: actions/upload-artifact@v4
        with:
          name: coverage-report
          path: coverage.out
```

## Advanced: Multiple Projects / Groups

For monorepos with multiple services:

```yaml
name: Multi-Project Coverage

on:
  push:
    branches: [main]

jobs:
  coverage:
    name: Coverage for ${{ matrix.project.name }}
    runs-on: ubuntu-latest
    strategy:
      matrix:
        project:
          - key: github.com/myorg/repo/api
            name: api
            group: backend
            path: ./services/api
          - key: github.com/myorg/repo/cli
            name: cli
            group: backend
            path: ./services/cli
          - key: github.com/myorg/repo/web
            name: web
            group: frontend
            path: ./web

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'

      - name: Test ${{ matrix.project.name }}
        run: |
          cd ${{ matrix.project.path }}
          go test ./... -coverprofile=coverage.out

      - name: Install coverage CLI
        run: go install github.com/arxdsilva/opencoverage/cmd/coveragecli@latest

      - name: Upload ${{ matrix.project.name }} coverage
        run: |
          cd ${{ matrix.project.path }}
          
          coveragecli \
            -coverprofile coverage.out \
            -out coverage-upload.json \
            -project-key "${{ matrix.project.key }}" \
            -project-name "${{ matrix.project.name }}" \
            -project-group "${{ matrix.project.group }}" \
            -branch "${{ github.ref_name }}" \
            -commit-sha "${{ github.sha }}" \
            -author "${{ github.actor }}" \
            -trigger-type "push"
          
          curl -X POST "${{ secrets.COVERAGE_API_URL }}" \
            -H "Content-Type: application/json" \
            -H "X-API-Key: ${{ secrets.COVERAGE_API_KEY }}" \
            --data-binary @coverage-upload.json
```

## Pull Request Integration

To validate coverage on pull requests:

```yaml
name: PR Coverage Check

on:
  pull_request:
    branches: [main]

env:
  COVERAGE_PROJECT_KEY: ${{ vars.COVERAGE_PROJECT_KEY }}
  COVERAGE_PROJECT_NAME: ${{ vars.COVERAGE_PROJECT_NAME }}
  COVERAGE_PROJECT_GROUP: ${{ vars.COVERAGE_PROJECT_GROUP }}

jobs:
  coverage:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'

      - name: Run tests with coverage
        run: go test ./... -coverprofile=coverage.out

      - name: Install coverage CLI
        run: go install github.com/arxdsilva/opencoverage/cmd/coveragecli@latest

      - name: Generate coverage payload
        run: |
          PROJECT_KEY="${COVERAGE_PROJECT_KEY:-${{ github.repository }}}"
          PROJECT_NAME="${COVERAGE_PROJECT_NAME:-${{ github.event.repository.name }}}"
          PROJECT_GROUP="${COVERAGE_PROJECT_GROUP:-backend}"

          coveragecli \
            -coverprofile coverage.out \
            -out coverage-upload.json \
            -project-key "$PROJECT_KEY" \
            -project-name "$PROJECT_NAME" \
            -project-group "$PROJECT_GROUP" \
            -branch "${{ github.head_ref }}" \
            -commit-sha "${{ github.sha }}" \
            -author "${{ github.actor }}" \
            -trigger-type "pr"

      - name: Upload to coverage-api
        id: coverage
        if: ${{ secrets.COVERAGE_API_URL != '' && secrets.COVERAGE_API_KEY != '' }}
        run: |
          RESPONSE=$(curl -sS -X POST "${{ secrets.COVERAGE_API_URL }}" \
            -H "Content-Type: application/json" \
            -H "X-API-Key: ${{ secrets.COVERAGE_API_KEY }}" \
            --data-binary @coverage-upload.json)
          
          echo "$RESPONSE" > coverage-response.json
          
          THRESHOLD=$(echo "$RESPONSE" | jq -r '.comparison.thresholdStatus')
          CURRENT=$(echo "$RESPONSE" | jq -r '.comparison.currentTotalCoveragePercent')
          DELTA=$(echo "$RESPONSE" | jq -r '.comparison.deltaPercent // "N/A"')
          
          echo "threshold=$THRESHOLD" >> $GITHUB_OUTPUT
          echo "current=$CURRENT" >> $GITHUB_OUTPUT
          echo "delta=$DELTA" >> $GITHUB_OUTPUT

      - name: Comment PR with results
        if: ${{ steps.coverage.outputs.threshold != null }}
        uses: actions/github-script@v7
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          script: |
            const fs = require('fs');
            const response = JSON.parse(fs.readFileSync('coverage-response.json', 'utf8'));
            const comparison = response.comparison;
            
            const emoji = comparison.thresholdStatus === 'passed' ? '✅' : '❌';
            const body = `
## Coverage Report ${emoji}

| Metric | Value |
|--------|-------|
| Current | ${comparison.currentTotalCoveragePercent.toFixed(2)}% |
| Previous | ${comparison.previousTotalCoveragePercent ? comparison.previousTotalCoveragePercent.toFixed(2) : 'N/A'}% |
| Delta | ${comparison.deltaPercent ? comparison.deltaPercent.toFixed(2) : 'N/A'}% |
| Threshold | ${comparison.thresholdPercent}% |
| Status | ${comparison.thresholdStatus} |
            `;
            
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: body
            });

      - name: Fail if coverage below threshold
        if: ${{ steps.coverage.outputs.threshold == 'failed' }}
        run: |
          echo "❌ Coverage is below threshold"
          exit 1
```

## Configuring The Workflow

The examples above make the project metadata configurable without editing the CLI command itself.

You can control these values in two ways:

1. Set repository variables:
   `COVERAGE_PROJECT_KEY`, `COVERAGE_PROJECT_NAME`, `COVERAGE_PROJECT_GROUP`
2. Edit the workflow-level `env:` block directly.

If you do not want to send a group, remove the `-project-group "$PROJECT_GROUP"` flag from the workflow step.

## CLI Installation Tips

### Latest Version
```yaml
- name: Install coverage CLI
  run: go install github.com/arxdsilva/opencoverage/cmd/coveragecli@latest
```

### Specific Version (Recommended for CI)
```yaml
- name: Install coverage CLI
  run: go install github.com/arxdsilva/opencoverage/cmd/coveragecli@v1.0.0
```

### From Local Source
```yaml
- name: Build coverage CLI
  run: go build -o /usr/local/bin/coveragecli ./cmd/coveragecli
```

## Environment Protection

To ensure coverage uploads only work with proper configuration:

```yaml
- name: Validate secrets are configured
  run: |
    if [ -z "${{ secrets.COVERAGE_API_URL }}" ] || [ -z "${{ secrets.COVERAGE_API_KEY }}" ]; then
      echo "⚠️  coverage-api secrets not configured, skipping upload"
      exit 0
    fi
```

## Troubleshooting

### Connection Refused
```
error: Failed to connect to http://your-server:8080/v1/coverage-runs
```
- Verify `COVERAGE_API_URL` secret is correct
- Ensure coverage-api is running and accessible from GitHub
- Check firewall rules allow outbound HTTPS/HTTP from GitHub runners
- For private networks, use a VPN or expose with reverse proxy

### Unauthorized (401)
```
error: Unauthorized - invalid API key
```
- Verify `COVERAGE_API_KEY` matches the API's `API_KEY_SECRET`
- Check secret is passed correctly in curl header
- Ensure no trailing spaces in the secret value

### CLI Not Found
```
command not found: coveragecli
```
- Add explicit step to install: `go install github.com/arxdsilva/opencoverage/cmd/coveragecli@latest`
- Or build from source in the workflow

### Coverage File Not Found
```
error: coverage.out: no such file or directory
```
- Ensure `go test ./... -coverprofile=coverage.out` runs successfully
- Check test step didn't fail silently

## Example Flow Diagram

```
GitHub Actions Trigger
    ↓
Checkout Code
    ↓
Set up Go
    ↓
Run Tests → Generate coverage.out
    ↓
Install coveragecli
    ↓
Convert coverage.out → coverage-upload.json
    ↓
Upload JSON to coverage-api
    ↓
View in Dashboard
```

## Next Steps

- Monitor coverage trends in the self-hosted dashboard
- Add coverage thresholds to enforce minimum coverage
- Set up status checks to block PRs with coverage regressions
- Integrate with team notifications (Slack, etc.)
