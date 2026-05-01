# GitHub Org Insights Specification (API + Frontend)

## 1. Overview

### 1.1 Purpose

Define how coverage-api exposes GitHub organization-level development health insights focused on reviewer activity and stale pull requests.

### 1.2 Primary Outcome

Provide deterministic API responses that answer:

1. Who is reviewing PRs the most in this organization over a given period?
2. Which open PRs are hanging and need attention?
3. How can the frontend visualize this data without direct GitHub auth in the browser?

### 1.3 Source of Truth

GitHub organization and repository pull request/review data is fetched server-side by a background worker and persisted as PostgreSQL snapshots. API reads return the latest persisted snapshot.

## 2. Scope

### 2.1 In Scope (v1)

1. Organization-level reviewer leaderboard endpoint backed by persisted snapshots.
2. Organization-level hanging PRs endpoint.
3. Configurable time windows and limits for both endpoints.
4. Deterministic hanging classification logic.
5. API key authentication on all endpoints.
6. Worker-based periodic sync from GitHub into PostgreSQL snapshots.
7. Frontend-selectable reviewer windows limited to `30`, `60`, or `90` days.

### 2.2 Out of Scope (v1)

1. Writing back to GitHub (labels, comments, assignments, merge actions).
2. Team-level leaderboards or per-team SLA policies.
3. Historical trend charting UI beyond latest snapshot windows.
4. Per-user privacy controls.

## 3. Key Decisions and Assumptions

1. Browser clients do not call GitHub directly; frontend uses coverage-api only.
2. A dedicated worker reads GitHub data using a server-side token from environment.
3. API responses are snapshot-based and can lag GitHub by the worker sync interval.
4. Missing review events produce zero counts; no synthetic data is inferred.
5. Hanging PR logic is rule-based and transparent in the response payload.

## 4. High-Level Architecture

1. Application ports:
  - `GitHubOrgInsightsService` (worker read-from-GitHub)
  - `GitHubOrgInsightsRepository` (snapshot writes/reads)
2. Adapters:
  - `internal/adapters/github` for worker GitHub REST calls.
  - `internal/adapters/postgres` for snapshot persistence and retrieval.
3. New HTTP handlers under existing API layer:
   - org leaderboard
   - org hanging PRs
4. GitHub insights sync worker loop runs inside API runtime.
5. Dependency direction remains inward (hexagonal boundaries preserved).

## 5. Configuration

Required/optional environment variables:

1. `GITHUB_TOKEN` (required by insights worker)
2. `GITHUB_API_BASE_URL` (optional, default `https://api.github.com`)
3. `GITHUB_INSIGHTS_CACHE_TTL_SECONDS` (optional, default `60`)
4. `GITHUB_INSIGHTS_MAX_REPOS` (optional safety cap, default `200`)
5. `GITHUB_ORGS` (required, comma-separated organizations for sync)
6. `GITHUB_INSIGHTS_WINDOW_DAYS` (optional, supported values `30,60,90`, default `30,60,90`)
7. `GITHUB_INSIGHTS_SYNC_INTERVAL_SECONDS` (optional, default `3600`)

Validation rules:

1. Worker startup fails fast when `GITHUB_TOKEN` or `GITHUB_ORGS` is missing.
2. `GITHUB_INSIGHTS_CACHE_TTL_SECONDS` must be >= 0.
3. `GITHUB_INSIGHTS_MAX_REPOS` must be >= 1.
4. `GITHUB_INSIGHTS_WINDOW_DAYS` values outside `30`, `60`, `90` are ignored.

## 6. API Contract (v1)

### 6.1 Authentication

1. Header: `X-API-Key: <key>`
2. Required for all endpoints.

### 6.2 Endpoint Summary

1. `GET /v1/github/orgs/{org}/reviewers/leaderboard`
2. `GET /v1/github/orgs/{org}/pull-requests/hanging`

## 7. Detailed Endpoint Specs

### 7.1 GET /v1/github/orgs/{org}/reviewers/leaderboard

Returns ranked reviewer activity for one GitHub organization.

Query params:

1. `from` (optional ISO-8601 timestamp)
2. `to` (optional ISO-8601 timestamp)
3. `windowDays` (optional integer, default 30, used only when `from` and `to` are absent)
4. `limit` (optional integer, default 20, max 100)
5. `repo` (optional, repeatable; when absent, include all org repos subject to max repo cap)

Behavior:

1. Read latest reviewer snapshot for (`org`, `windowDays`) from PostgreSQL.
2. Return persisted reviewer ordering and apply `limit` over snapshot rows.
3. Return `404 NOT_FOUND` when no snapshot exists for (`org`, `windowDays`).

Response example:

```json
{
  "org": "acme",
  "window": {
    "from": "2026-04-01T00:00:00Z",
    "to": "2026-05-01T00:00:00Z"
  },
  "summary": {
    "repositoriesScanned": 34,
    "pullRequestsConsidered": 412,
    "totalReviewEvents": 987
  },
  "reviewers": [
    {
      "login": "alice",
      "displayName": "Alice Doe",
      "totalReviews": 82,
      "approvals": 49,
      "changeRequests": 11,
      "comments": 22,
      "uniquePullRequestsReviewed": 61,
      "reposReviewed": 12,
      "latestReviewAt": "2026-04-30T22:21:00Z"
    }
  ]
}
```

### 7.2 GET /v1/github/orgs/{org}/pull-requests/hanging

Returns open pull requests that are considered hanging.

Query params:

1. `limit` (optional integer, default 50, max 200)
2. `minIdleHours` (optional integer, default 48)
3. `minOpenHours` (optional integer, default 72)
4. `repo` (optional, repeatable)
5. `author` (optional exact login)
6. `includeDrafts` (optional boolean, default false)
7. `sort` (optional enum: `staleness_desc`, `open_time_desc`; default `staleness_desc`)

Hanging classification rules (worker-time, v1):

1. Worker snapshots only include open PRs that match hanging rules at sync time.
2. API request-time filters (`minOpenHours`, `minIdleHours`, `author`, `repo`, `includeDrafts`) are applied on the latest snapshot.
3. At least one of the following reasons must apply:
   - `awaiting-review`: no completed review exists.
   - `changes-requested`: latest review state is changes requested and no newer author push.
   - `awaiting-author`: latest review requests changes and author has pushed but PR is still idle beyond threshold.
   - `merge-conflict`: PR reports merge conflicts.

Response example:

```json
{
  "org": "acme",
  "generatedAt": "2026-05-01T00:00:00Z",
  "criteria": {
    "minIdleHours": 48,
    "minOpenHours": 72,
    "includeDrafts": false
  },
  "summary": {
    "repositoriesScanned": 34,
    "openPullRequestsConsidered": 517,
    "hangingPullRequests": 37
  },
  "items": [
    {
      "repository": "platform-api",
      "number": 1284,
      "title": "Refactor comparison baseline selection",
      "url": "https://github.com/acme/platform-api/pull/1284",
      "author": "devon",
      "draft": false,
      "createdAt": "2026-04-20T14:02:00Z",
      "updatedAt": "2026-04-26T09:11:00Z",
      "lastActivityAt": "2026-04-26T09:11:00Z",
      "ageHours": 252,
      "idleHours": 111,
      "reviewState": "changes_requested",
      "mergeableState": "dirty",
      "requestedReviewers": ["alice", "maria"],
      "labels": ["backend", "api"],
      "reasons": ["changes-requested", "merge-conflict"]
    }
  ]
}
```

## 8. Error Model

Same structured error model used in `specs/SPEC.md`:

1. `UNAUTHENTICATED` (401)
2. `INVALID_ARGUMENT` (400)
3. `NOT_FOUND` (404) for unknown org/repo in configured provider context
4. `RATE_LIMITED` (429) when GitHub API limit is exceeded
5. `INTERNAL` (500)

Error payload shape:

```json
{
  "error": {
    "code": "RATE_LIMITED",
    "message": "GitHub API rate limit exceeded",
    "details": {
      "provider": "github"
    }
  }
}
```

## 9. Validation Rules

1. `org` path param required and non-empty.
2. `windowDays` allowed values on frontend: `30`, `60`, `90`.
3. `limit` range:
   - leaderboard: 1 to 100
   - hanging PRs: 1 to 200
4. `from` must be <= `to` when both present.
5. `minIdleHours` and `minOpenHours` must be >= 1.
6. If `repo` is provided, each value must be unique.

## 10. Sync and Freshness Strategy

1. Worker syncs each configured org periodically (default every 1 hour).
2. Reviewer snapshots are persisted per (`org`, `windowDays`) for supported windows `30`, `60`, `90`.
3. Hanging PR snapshots are persisted per org and read as latest snapshot.
4. API responses are deterministic for latest persisted snapshot state.

## 11. Dedicated Org Insights Screen (Product Specification)

### 11.1 Product Intent

The Org Insights experience is a dedicated screen, not an overlay and not a panel inside Home.

Goals:

1. Create a single place where engineering leaders can identify review bottlenecks.
2. Show which reviewers are carrying the load.
3. Surface PRs that are stuck with clear reason codes.
4. Reduce time-to-action from insight to GitHub follow-up.

Non-goals:

1. Replacing GitHub PR details pages.
2. Performing workflow automation in v1.
3. Offering historical trend analytics in v1.

### 11.2 Route, Navigation, and URL Contract

1. New frontend route: `/org-insights`.
2. Sidebar must include `Org Insights` as a primary navigation item.
3. The screen must support deep-linking via query params:

    - `org`
    - `windowDays`
    - `from` (API-supported, not exposed by v1 frontend controls)
    - `to` (API-supported, not exposed by v1 frontend controls)
    - `repo` (repeatable)
    - `limit`
    - `minIdleHours`
    - `minOpenHours`
    - `includeDrafts`
    - `author`
    - `sort`
4. Navigation between `/`, `/integration`, and `/org-insights` must preserve browser back/forward behavior.
5. Query parameter state must be reflected in UI controls on page load.

### 11.3 Primary Users and Jobs-to-be-Done

1. Engineering manager:
    - Need to identify review load imbalance and stalled work.
    - Need a list of PRs to triage immediately.
2. Tech lead:
    - Need to monitor team throughput and review discipline.
    - Need to quickly find PRs blocked by merge conflicts or stale change requests.
3. Individual contributor:
    - Need visibility into where to help with reviews.

### 11.4 Information Architecture

The screen is composed of six regions:

1. Header bar.
2. Global filter bar.
3. KPI summary strip.
4. Reviewer leaderboard module.
5. Hanging PR queue module.
6. Context drawer for selected PR.

### 11.5 Header Bar

Required controls and content:

1. Title: `Org Insights`.
2. Subtitle: current org and active time window.
3. `Refresh` button.
4. Auto-refresh selector (`off`, `15s`, `30s`, `60s`, `5m`).
5. Last updated timestamp label.

Header behavior:

1. `Refresh` triggers both leaderboard and hanging PR requests in parallel.
2. During refresh, show in-progress state and disable duplicate clicks.
3. Last updated timestamp uses API response `generatedAt` when available, else client receipt timestamp.

### 11.6 Global Filter Bar

Required filters:

1. Organization selector/input (required).
2. Relative `windowDays` selector with allowed values `30`, `60`, `90`.
3. Repository multi-select.
4. Reviewer leaderboard limit.
5. Hanging queue controls:
    - `minIdleHours`
    - `minOpenHours`
    - `includeDrafts`
    - `author` filter
    - `sort`

Filter interaction rules:

1. `Apply` updates URL and triggers refresh.
2. `Reset` restores sensible defaults and clears URL query params except `org`.
3. Invalid filter values must render inline field errors before any request is made.
4. Active filter chips must be visible and removable one by one.

### 11.7 KPI Summary Strip

Show at least these cards:

1. `Hanging PRs` (count).
2. `Open PRs Considered` (count).
3. `Review Events` (count in selected window).
4. `Top Reviewer` (login + total reviews).

Behavior:

1. KPI cards update after both module queries resolve.
2. Partial failures render per-card fallback (`--`) with tooltip reason.
3. Clicking `Hanging PRs` card scrolls/focuses hanging queue module.

### 11.8 Reviewer Leaderboard Module

Purpose:

1. Rank reviewers by completed review workload and quality signal.

Presentation:

1. Table on desktop.
2. Card list on narrow screens.

Columns/fields:

1. Rank
2. Reviewer (avatar, display name, login)
3. Total reviews
4. Approvals
5. Change requests
6. Comments
7. Unique PRs reviewed
8. Repos reviewed
9. Latest review at

Interactions:

1. Sort control supports:
    - default rank sort from API
    - approvals desc
    - change requests desc
2. Clicking a reviewer applies a local highlight and filters hanging queue by reviewer involvement (requested reviewer or prior reviewer).
3. Export control offers CSV download of currently visible rows.

Empty and edge states:

1. If no data in window, show `No review activity found for current filters.`
2. If API partially fails, keep last successful table and show non-blocking warning banner.

### 11.9 Hanging PR Queue Module

Purpose:

1. Provide action-oriented list of open PRs that require intervention.

Presentation:

1. Dense table on desktop.
2. Expandable cards on mobile.

Columns/fields:

1. Repository/name
2. PR number and title
3. Author
4. Reason chips (one or many)
5. Review state
6. Idle time
7. Open age
8. Requested reviewers
9. Labels
10. Actions (`Open in GitHub`, `Copy Link`)

Reason chip definitions:

1. `awaiting-review`: no completed review exists.
2. `changes-requested`: requested changes not addressed in review flow.
3. `awaiting-author`: author pushed after requested changes but PR remains stale.
4. `merge-conflict`: PR has merge conflicts.

Interactions:

1. Clicking a row opens context drawer.
2. Multi-select reasons filters list client-side.
3. `Only my authored PRs` toggle applies local author = current input value.
4. Pagination or incremental load supports at least 200 results.

### 11.10 PR Context Drawer

Drawer content:

1. PR title, number, repo, URL.
2. Timeline summary:
    - created
    - last update
    - last review event
    - last author push (when available)
3. Hanging reasons with plain-language explanation.
4. Requested reviewers and labels.
5. Suggested next actions checklist (non-automated guidance text):
    - ping reviewers
    - resolve conflicts
    - re-request review

Drawer behavior:

1. Opens from row click.
2. Closes on escape key and close button.
3. Keeps current filter context intact.

### 11.11 Loading, Empty, Error, and Partial-Success States

1. Initial load:
    - show screen skeleton for all modules.
2. Refresh load:
    - keep existing data visible and apply top progress indicator.
3. Leaderboard failure only:
    - show module-level error, keep hanging PR module functional.
4. Hanging PR failure only:
    - show module-level error, keep leaderboard functional.
5. Full failure:
    - show page-level error with `Retry` action and preserve current filter controls.
6. Empty org/repo result:
    - explicit empty message, not error styling.

### 11.12 Refresh and Auto-Refresh Behavior

1. Manual refresh fetches both modules in parallel.
2. Auto-refresh uses same semantics as existing screens, with Org Insights defaulting to a longer cadence.
3. Org Insights auto-refresh is enabled by default at `1h` when no prior user preference exists.
4. If a refresh is in flight when interval fires, skip overlapping refresh.
5. Pause polling when document is hidden; resume on visible.
6. Status line requirements:
    - `Auto refresh is off.`
    - `Refreshing now (...)`
    - `Paused (...) while tab is hidden.`
    - `Next refresh in Ns`

### 11.13 State Persistence

Persist per route (`/org-insights`) using local storage:

1. Last selected org.
2. Filter selections.
3. Column visibility preferences.
4. Auto-refresh interval.
5. Last selected module sort option.

### 11.14 Accessibility and UX Quality Bar

1. Keyboard navigation across all interactive controls.
2. Focus-visible states on all controls.
3. Tables/cards support screen-reader labels for reason chips and action buttons.
4. Color is not the only signal for hanging reasons; always include text labels.
5. Mobile layout remains usable at 360px width.

### 11.15 Performance and Reliability Requirements

1. Target P95 time-to-first-data for cached responses <= 1.0s.
2. Target P95 API response time <= 1.5s for snapshot reads.
3. Worker sync duration and failure rate must be monitored independently from API response latency.
4. Frontend must render first meaningful content with skeletons immediately.
5. Screen must tolerate API partial failures without full-page reset.

### 11.16 Telemetry and Product Analytics

Track these client-side events:

1. `org_insights_screen_viewed`
2. `org_insights_filters_applied`
3. `org_insights_refresh_clicked`
4. `org_insights_auto_refresh_changed`
5. `org_insights_reviewer_clicked`
6. `org_insights_hanging_pr_row_opened`
7. `org_insights_export_csv`

Track these server-side counters/histograms:

1. Request count by endpoint/org.
2. Snapshot read latency.
3. Worker sync duration.
4. Worker sync failure count by org.
5. Rate-limit error count.

### 11.17 Frontend API Proxy Routes

Browser-facing routes (frontend server):

1. `GET /api/github/orgs/{org}/reviewers/leaderboard`
2. `GET /api/github/orgs/{org}/pull-requests/hanging`

Proxied API routes (coverage-api):

1. `GET /v1/github/orgs/{org}/reviewers/leaderboard`
2. `GET /v1/github/orgs/{org}/pull-requests/hanging`

### 11.18 Rollout Plan

1. Feature-flag the route and sidebar entry.
2. Internal alpha with one org and limited repos.
3. Validate hanging classification feedback with engineering leads.
4. Gradual enablement by environment.

## 12. Testing Strategy

1. Unit tests for hanging classification rules with table-driven cases.
2. Unit tests for leaderboard aggregation and ordering tie-breakers.
3. Use case tests with mocked `GitHubOrgInsightsService` port and cache behavior.
4. HTTP handler tests for validation, query parsing, and error mapping.
5. Frontend tests for filter persistence, URL sync, and partial-failure rendering.
6. Frontend tests for auto-refresh pause/resume and countdown states.
7. Adapter integration tests against GitHub API contract fixtures.

## 13. Acceptance Criteria

API criteria:

1. `GET /v1/github/orgs/{org}/reviewers/leaderboard` returns deterministic sorted reviewer metrics.
2. `GET /v1/github/orgs/{org}/pull-requests/hanging` returns snapshot-backed hanging PRs with request-time filtering.
3. Invalid query parameters return `400 INVALID_ARGUMENT` with field details.
4. GitHub rate-limit scenarios map to `429 RATE_LIMITED`.
5. API key auth is enforced consistently.

Frontend screen criteria:

1. Org Insights exists as dedicated route `/org-insights`.
2. Screen shows independent modules for leaderboard and hanging PR queue.
3. Filters are deep-linkable via URL query params and restored on load.
4. Frontend window selector only allows `30`, `60`, or `90` day options.
5. Manual refresh and auto-refresh follow shared refresh semantics.
6. Org Insights auto-refresh defaults to `1h` unless user has an existing saved preference.
7. Module-level failures do not blank the entire screen.
8. User can open PR context drawer and navigate to GitHub from each row.
9. Frontend renders both datasets without direct GitHub API access.

## 14. Future Enhancements (Post-v1)

1. Team-aware leaderboards and ownership weighting.
2. SLA policy profiles per repository.
3. Historical trend persistence in PostgreSQL.
4. Auto-generated weekly digest endpoint.
5. Optional webhook events for newly hanging PRs.
6. Smart recommendations for reviewer rebalancing.
