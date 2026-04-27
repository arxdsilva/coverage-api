-- +goose Up
CREATE TYPE environment_type AS ENUM ('test', 'stage', 'prod');

CREATE TABLE IF NOT EXISTS integration_test_runs (
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL REFERENCES projects(id),
  branch TEXT NOT NULL,
  commit_sha TEXT NOT NULL,
  author TEXT,
  trigger_type TEXT NOT NULL CHECK (trigger_type IN ('push', 'pr', 'manual')),
  run_timestamp TIMESTAMPTZ NOT NULL,
  ginkgo_version TEXT,
  suite_description TEXT NOT NULL,
  suite_path TEXT NOT NULL,
  total_specs INTEGER NOT NULL,
  passed_specs INTEGER NOT NULL,
  failed_specs INTEGER NOT NULL,
  skipped_specs INTEGER NOT NULL,
  flaked_specs INTEGER NOT NULL,
  pending_specs INTEGER NOT NULL,
  interrupted BOOLEAN NOT NULL DEFAULT FALSE,
  timed_out BOOLEAN NOT NULL DEFAULT FALSE,
  duration_ms BIGINT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('passed', 'failed')),
  environment environment_type,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS integration_runs_project_branch_ts_idx
  ON integration_test_runs(project_id, branch, run_timestamp DESC);

CREATE INDEX IF NOT EXISTS integration_runs_project_default_lookup_idx
  ON integration_test_runs(project_id, run_timestamp DESC);

CREATE INDEX IF NOT EXISTS integration_runs_project_status_ts_idx
  ON integration_test_runs(project_id, status, run_timestamp DESC);

CREATE TABLE IF NOT EXISTS integration_spec_results (
  id UUID PRIMARY KEY,
  integration_run_id UUID NOT NULL REFERENCES integration_test_runs(id) ON DELETE CASCADE,
  spec_path TEXT NOT NULL,
  leaf_node_text TEXT NOT NULL,
  state TEXT NOT NULL CHECK (state IN ('passed', 'failed', 'skipped', 'pending', 'flaky')),
  duration_ms BIGINT NOT NULL,
  failure_message TEXT,
  failure_location_file TEXT,
  failure_location_line INTEGER
);

CREATE INDEX IF NOT EXISTS integration_spec_results_run_id_idx ON integration_spec_results(integration_run_id);
CREATE INDEX IF NOT EXISTS integration_spec_results_state_idx ON integration_spec_results(state);

-- +goose Down
DROP TABLE IF EXISTS integration_spec_results;
DROP TABLE IF EXISTS integration_test_runs;
DROP TYPE IF EXISTS environment_type;
