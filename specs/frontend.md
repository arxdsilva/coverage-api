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

## Visual Refresh

The frontend now uses a light, modern visual direction inspired by QuintoAndar.

Implemented style goals:

1. Move to a light surface system with white cards and soft neutral backgrounds.
2. Keep strong readability using dark text and restrained accent colors.
3. Modernize controls (buttons, inputs, selects) with lighter borders, clear focus states, and cleaner spacing.
4. Preserve all current information architecture and interactions while only changing visual language.
5. Keep responsive behavior and route-level separation (`/` and `/integration`) unchanged.

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

## URL Query Parameters

Both the home page and integration screen support opening overlays via URL query parameters:

### `?heatmap=open`

Opens the heatmap overlay on page load.

**Examples:**

- Home page with heatmap open: `http://localhost:8090/?heatmap=open`
- Integration screen with heatmap open: `http://localhost:8090/integration?heatmap=open`

This allows deep linking to either heatmap visualization directly.

Home route behavior:

1. The left sidebar `Select Project` area supports two-step selection: choose a project group first, then choose a project from that filtered group.
2. Group filtering and project-name search are combined. The project dropdown only lists projects matching both filters.
3. If current filters exclude the selected project, the screen automatically selects the first matching project; if no project matches, the screen enters a filtered empty state without clearing filter controls.

Integration route behavior:

1. The integration dashboard includes a `Run Chain` graphic rendered from the current run history query.
2. Each node in the chain represents one integration run.
3. Node color encodes run status: green for `passed`, red for `failed`.
4. Nodes are ordered oldest to newest, connected by arrows, with newest on the right.
5. Selecting a node focuses the corresponding run and refreshes failed-spec details.
6. Run Chain shows at most 5 runs (newest 5 from the current run-list query).
7. The integration dashboard includes an `Integration Heatmap` that shows recent runs for **all projects** simultaneously, organized by project group.
8. Heatmap layout: projects are grouped by their `group` field, rendered as labeled sections. Within each group, one row per project, one tile per run ordered oldest-to-newest (newest on the right). Projects with no group appear in an unlabeled section at the bottom.
9. Heatmap data is sourced from `GET /api/integration-test-runs/heatmap` (all-project aggregated endpoint), which returns runs pre-grouped by project group.
10. Heatmap displays only each project's default-branch runs.
11. Heatmap tiles use emoji status markers: `✅` for pass and `❌` for fail.
12. Heatmap row tint reflects the newest run result for that project (passed/failed).
13. Clicking a heatmap tile for the active project synchronizes selected state with run chain, run table, and failed-spec details.
14. Heatmap supports status filtering and local reload; branch filter is fixed to default-branch-only behavior.
15. The left sidebar `Select Project` area supports two-step selection: choose a project group first, then choose a project from that filtered group.
16. Group filtering and project-name search are combined. The project dropdown only lists projects matching both filters.
17. If current filters exclude the selected project, the screen automatically selects the first matching project; if no project matches, the screen enters a filtered empty state without clearing filter controls.

## Refresh Mechanism

The frontend must support explicit manual refresh on both primary screens.

### Goals

1. Refresh must give users a predictable way to pull newly ingested data without a full browser reload.
2. Refresh must preserve as much UI context as possible: current route, selected project, active overlays, query-parameter-driven overlay state, and user-entered filters.
3. Refresh must keep failures localized. A failed sub-request must not blank unrelated panels when existing data can still be shown.
4. Refresh must support configurable automatic polling so users can keep either screen current without manually reloading.

### Shared Rules

1. Each screen must expose a primary `Refresh` control in the sidebar or screen header.
2. A refresh action must be idempotent and safe to trigger repeatedly.
3. While a refresh is in flight, the triggering control should enter a loading/disabled state to prevent duplicate concurrent refreshes from the same control.
4. Refresh must not perform a full document reload.
5. Refresh completion should update only the affected data views and should not collapse open overlays or reset the sidebar state.
6. If the currently selected project still exists after refresh, it must remain selected.
7. If the currently selected project no longer exists, the UI should fall back to the first available project; if no projects remain, the screen should enter its empty state.
8. Search text used only to filter the visible project selector may be cleared on refresh, but persisted screen-level filters must be preserved unless the server data makes them invalid.
9. Query-parameter-opened overlays, such as `?heatmap=open`, must remain open after refresh unless the user explicitly closes them.
10. Errors must be surfaced inline near the affected panel or overlay, using the last successfully rendered data when practical.

### Auto Refresh Configuration

1. Each screen must expose an auto-refresh control alongside the manual refresh action.
2. The control must allow choosing one of these intervals: `off`, `15s`, `30s`, `60s`, `5m`.
3. `off` disables automatic polling; any other value enables periodic refresh using the selected interval.
4. The selected interval must be configurable independently per route (`/` and `/integration`).
5. The selected interval should persist across reloads for the same route, for example via browser local storage.
6. A manual refresh and an automatic refresh must execute the same underlying refresh workflow for that screen.
7. If a refresh is already running when the next interval fires, the overlapping automatic refresh must be skipped rather than queued.
8. A successful or failed manual refresh should reset the auto-refresh timer so the next automatic tick is scheduled from the end of the most recent refresh attempt.
9. Automatic refresh should pause while the document is hidden and resume when the tab becomes visible again.
10. Automatic refresh must preserve the same state guarantees as manual refresh: selected project, selected run, overlay open state, filters, and query-parameter-driven overlay state.
11. Automatic refresh must not steal focus, scroll the page unexpectedly, or close overlays.
12. Each screen must display a visible auto-refresh status line near the interval selector.
13. The status line must show the next automatic refresh timing as a live countdown while auto refresh is enabled (for example `Next refresh in 12s`).
14. The status line must switch to explicit state labels when applicable:
	- `Auto refresh is off.` when interval is `off`
	- `Refreshing now (...)` while a refresh is in flight
	- `Paused (...) while tab is hidden.` when polling is paused by document visibility
15. Countdown/status text must update automatically without requiring manual user interaction.

### Home Screen Refresh Scope

1. Trigger: the sidebar `Refresh` button on `/`.
2. A home-screen refresh must refetch the project catalog first.
3. After the project catalog is refreshed, the home screen must refresh the active project's dependent views:
	- branch list
	- recent runs table
	- latest comparison summary
	- multi-branch trend chart
4. The all-project coverage heatmap dataset must also be refreshed as part of the same action.
5. If the heatmap overlay is currently open, it must remain open and repaint with the refreshed dataset.
6. If the Top Contributors overlay is currently open, it should remain open and refresh its contributor dataset as part of the same action.
7. A project-selector change is not equivalent to a global refresh; it refreshes only the newly selected project's dependent views plus any views that are defined as global for the screen.
8. When home-screen auto refresh is enabled, each automatic interval must run this same screen-level refresh scope.

### Integration Screen Refresh Scope

1. Trigger: the sidebar `Refresh` button on `/integration`.
2. An integration-screen refresh must refetch the project catalog first.
3. After the project catalog is refreshed, the screen must refresh the active project's dependent views:
	- latest integration comparison summary
	- integration run table
	- run chain visualization
	- failed-spec details for the selected run
4. If the selected run no longer exists after refresh, the screen should select the newest available run from the refreshed run list.
5. The integration heatmap overlay dataset is global to the screen and should refresh as part of the same action when the overlay is open.
6. The integration heatmap overlay must also support its own local `Reload` control that refreshes only the heatmap query and preserves overlay filters.
7. The per-project integration table `Reload` control must refresh only the active project's comparison, run list, run chain, and failed-spec details; it must not implicitly reset heatmap filters.
8. When integration-screen auto refresh is enabled, each automatic interval must run this same screen-level refresh scope.

### Loading and Error Behavior

1. Refreshing a panel should preserve the previous rendered state until replacement data arrives, unless an empty/loading placeholder is materially clearer.
2. Overlay-specific reload actions must show loading state inside the overlay body rather than in unrelated screen panels.
3. A project-catalog refresh failure must not clear already rendered project data unless the UI can no longer determine a valid selection.
4. A global refresh may complete partially. The spec allows mixed success states as long as each affected region communicates its own failure clearly.
5. Automatic refresh failures should be surfaced non-destructively and must not disable future scheduled refresh attempts unless the user explicitly turns auto refresh off.
6. Auto-refresh status text must remain truthful during errors (for example, keep countdown behavior and interval context after a failed automatic tick).

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
