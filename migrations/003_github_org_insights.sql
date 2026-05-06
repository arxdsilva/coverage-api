-- +goose Up
CREATE TABLE IF NOT EXISTS github_reviewer_snapshots (
  id UUID PRIMARY KEY,
  org_name TEXT NOT NULL,
  window_days INT NOT NULL,
  window_from TIMESTAMPTZ NOT NULL,
  window_to TIMESTAMPTZ NOT NULL,
  repositories_scanned INT NOT NULL,
  pull_requests_considered INT NOT NULL,
  total_review_events INT NOT NULL,
  generated_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS github_reviewer_snapshots_org_window_idx
  ON github_reviewer_snapshots(org_name, window_days, generated_at DESC);

CREATE TABLE IF NOT EXISTS github_reviewer_entries (
  id UUID PRIMARY KEY,
  snapshot_id UUID NOT NULL REFERENCES github_reviewer_snapshots(id) ON DELETE CASCADE,
  login TEXT NOT NULL,
  display_name TEXT,
  total_reviews INT NOT NULL,
  approvals INT NOT NULL,
  change_requests INT NOT NULL,
  comments INT NOT NULL,
  unique_pull_requests_reviewed INT NOT NULL,
  repos_reviewed INT NOT NULL,
  latest_review_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS github_reviewer_entries_snapshot_idx
  ON github_reviewer_entries(snapshot_id);

CREATE TABLE IF NOT EXISTS github_hanging_snapshots (
  id UUID PRIMARY KEY,
  org_name TEXT NOT NULL,
  repositories_scanned INT NOT NULL,
  open_pull_requests_considered INT NOT NULL,
  hanging_pull_requests INT NOT NULL,
  generated_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS github_hanging_snapshots_org_idx
  ON github_hanging_snapshots(org_name, generated_at DESC);

CREATE TABLE IF NOT EXISTS github_hanging_items (
  id UUID PRIMARY KEY,
  snapshot_id UUID NOT NULL REFERENCES github_hanging_snapshots(id) ON DELETE CASCADE,
  repository_name TEXT NOT NULL,
  pr_number INT NOT NULL,
  title TEXT NOT NULL,
  url TEXT NOT NULL,
  author TEXT NOT NULL,
  draft BOOLEAN NOT NULL,
  created_at_pr TIMESTAMPTZ NOT NULL,
  updated_at_pr TIMESTAMPTZ NOT NULL,
  last_activity_at TIMESTAMPTZ NOT NULL,
  age_hours INT NOT NULL,
  idle_hours INT NOT NULL,
  review_state TEXT NOT NULL,
  mergeable_state TEXT NOT NULL,
  requested_reviewers TEXT[] NOT NULL,
  labels TEXT[] NOT NULL,
  reasons TEXT[] NOT NULL
);

CREATE INDEX IF NOT EXISTS github_hanging_items_snapshot_idx
  ON github_hanging_items(snapshot_id);

-- +goose Down
DROP TABLE IF EXISTS github_hanging_items;
DROP TABLE IF EXISTS github_hanging_snapshots;
DROP TABLE IF EXISTS github_reviewer_entries;
DROP TABLE IF EXISTS github_reviewer_snapshots;