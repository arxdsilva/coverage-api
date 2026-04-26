# Frontend Guide

## Overview

This repository includes a separate frontend entrypoint at `cmd/frontend`.

The frontend is a dark-theme dashboard inspired by observability UIs (SigNoz-like):

- project list
- multi-branch coverage trend chart
- latest comparison summary
- package-level comparison table
- recent runs table

## Coverage Trend

The overview card includes a `Coverage Trend` graph for the selected project.

Current behavior:

- always shows the default branch plus every discovered project branch on the same graph
- loads up to `10` recent runs per branch
- uses run time on the X axis rather than commit SHA labels
- highlights the default branch separately so it remains easy to compare against feature branches
- updates when the selected project changes

The branch selector below the overview card does **not** filter the trend graph. It controls the latest comparison panel and package comparison table only.

## Integration Screen Navigation

Integration test views are available on a dedicated frontend route:

- `/integration`

Navigation behavior:

1. The main dashboard includes an `Integration Tests` control in the sidebar that navigates to `/integration`.
2. The integration screen includes a `Back to Home` control that navigates to `/`.
3. Home and integration are fully separate frontend documents and scripts:
	- `/` serves `index.html` with `assets/app.js`
	- `/integration` serves `integration.html` with `assets/integration.js`
4. Browser back/forward navigation between `/` and `/integration` is supported.
5. Navigation works even when JavaScript fails to initialize because these controls are real links (`<a href=...>`).

Integration route behavior:

1. The integration dashboard includes a `Run Chain` graphic rendered from the current run history query.
2. Each node in the chain represents one integration run.
3. Node color encodes run status: green for `passed`, red for `failed`.
4. Nodes are ordered newest to oldest and connected in a horizontal chain.
5. Selecting a node focuses the corresponding run and refreshes failed-spec details.

## No User Authentication

The frontend does not require user login/auth.

How it works:

1. Browser calls frontend endpoints under `/api/*`.
2. Frontend server proxies to coverage-api (`/v1/*`).
3. Frontend injects `API_KEY_SECRET` server-side into the API key header.
4. API key is never entered by the user in the UI.

## Folder Structure

- `cmd/frontend/main.go` - frontend server and API proxy
- `cmd/frontend/web/index.html` - coverage dashboard shell
- `cmd/frontend/web/integration.html` - integration dashboard shell
- `cmd/frontend/web/assets/styles.css` - dark theme styles
- `cmd/frontend/web/assets/app.js` - home coverage UI logic and API calls
- `cmd/frontend/web/assets/integration.js` - integration UI logic and API calls

## Configuration

Environment variables used by frontend:

- `FRONTEND_ADDR` (default `:8090`)
- `API_BASE_URL` (default `http://localhost:8080`)
- `API_KEY_HEADER` (default `X-API-Key`)
- `API_KEY_SECRET` (default `dev-local-key`)

## Run

### Run API

```bash
export DATABASE_URL="postgres://coverage:coverage@localhost:5433/coverage?sslmode=disable"
export API_KEY_SECRET="dev-local-key"
go run ./cmd/api
```

### Run Frontend

```bash
FRONTEND_ADDR=":8090" API_BASE_URL="http://localhost:8080" API_KEY_SECRET="dev-local-key" go run ./cmd/frontend
```

Open:

- `http://localhost:8090`

### Run With Docker Compose

```bash
make compose-up
```

Services started by compose:

- API on `http://localhost:8080`
- Frontend on `http://localhost:8090`

## Make Targets

- `make frontend-run`
- `make frontend-dev`

`frontend-dev` prints the two commands to run API and frontend in separate terminals.

## Heatmap and Group Visualization

The dashboard includes an interactive heatmap view of all projects accessible via the "Open Heatmap" button.

### Features

- **Group Organization**: Projects with an assigned group appear as balanced panels within the heatmap overlay
- **Responsive Layout**: Groups automatically reflow to fill the visible overlay area as panels
- **Color Coding**: Individual project tiles keep the existing delta coloring model and CSS classes. The only change is the delta source: it is now computed from the two most recent commits on the `main` branch only. The current seven classes (`delta-neg-3` through `delta-pos-3`) still map onto the same -3% to +3% red-to-green gradient — red shades for regression, neutral grey for no change, green shades for improvement. A legend strip in the overlay header shows the full scale at a glance.
- **Per-Group Tiles**: Each group contains project tiles sized to fill the group's allocated space
- **Real-time Relayout**: Groups reflow on window resize to maintain optimal use of screen space

### Delta Color Scale

Project tile background colors keep the same thresholds and visual meanings as before, but the delta input is now the difference between the latest `main` branch commit and the immediately previous `main` branch commit. The heatmap does not compare against feature branches, pull request branches, or any older baseline outside that most recent two-commit window.

| Class | Delta range | Color |
|---|---|---|
| `delta-pos-3` | +3% or more | Deep green |
| `delta-pos-2` | +2% | Green |
| `delta-pos-1` | +1% | Light green |
| `delta-zero` | 0% | Neutral grey-blue |
| `delta-neg-1` | -1% | Light red |
| `delta-neg-2` | -2% | Red |
| `delta-neg-3` | -3% or worse | Deep red |

Projects without at least two `main` branch commits to compare receive `delta-zero` styling. Group container backgrounds still use a subtle green/red tint based on the same latest-vs-previous `main` branch comparison (`heatmap-group-up` / `heatmap-group-down`).

Ungrouped projects appear in an "Ungrouped" panel at the bottom of the heatmap.

## Top Contributors Overlay

The dashboard includes a global Top Contributors view accessible via the "Top Contributors" button.

### Features

- **Global View**: Shows contributor rankings across all projects simultaneously — no project selection required.
- **Grouped Layout**: Projects are grouped by the same group name used in the heatmap. Groups are sorted alphabetically; ungrouped projects appear last.
- **Default Branch Scope**: Contributor stats are computed only from runs on each project's configured default branch.
- **Per-Project Blocks**: Each project shows its name, default branch, and an ordered list of contributors with commit count, run count, average coverage, and latest run timestamp.
- **Parallel Fetch**: The frontend fetches contributor data for all projects concurrently and renders when all data arrives.
- **Dynamic Full-Screen Layout**: Uses the same `fitGrid` layout engine as the heatmap — groups fill the entire overlay area with columns and rows computed based on aspect ratio. Relayouts on window resize.
- **Limit**: Each project shows its top 5 contributors by default (configurable via `limit` query param, server max 25).

## Exposed Frontend API Proxy Routes

The frontend server exposes unauthenticated GET routes for browser use:

- `GET /api/projects`
- `GET /api/projects/{projectId}/branches`
- `GET /api/projects/{projectId}/coverage-runs`
- `GET /api/projects/{projectId}/coverage-runs/latest-comparison`
- `GET /api/projects/{projectId}/contributors`

These are proxied to:

- `GET /v1/projects`
- `GET /v1/projects/{projectId}/branches`
- `GET /v1/projects/{projectId}/coverage-runs`
- `GET /v1/projects/{projectId}/coverage-runs/latest-comparison`
- `GET /v1/projects/{projectId}/contributors`
