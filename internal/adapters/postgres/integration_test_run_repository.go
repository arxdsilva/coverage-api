package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arxdsilva/opencoverage/internal/application"
	"github.com/arxdsilva/opencoverage/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IntegrationTestRunRepository struct {
	pool *pgxpool.Pool
}

func NewIntegrationTestRunRepository(pool *pgxpool.Pool) *IntegrationTestRunRepository {
	return &IntegrationTestRunRepository{pool: pool}
}

func (r *IntegrationTestRunRepository) Create(ctx context.Context, run domain.IntegrationTestRun) (domain.IntegrationTestRun, error) {
	q := getQuerier(ctx, r.pool)
	_, err := q.Exec(ctx, `
		INSERT INTO integration_test_runs (
			id, project_id, branch, commit_sha, author, trigger_type, run_timestamp,
			ginkgo_version, suite_description, suite_path, total_specs, passed_specs,
			failed_specs, skipped_specs, flaked_specs, pending_specs, interrupted,
			timed_out, duration_ms, status, environment, created_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22
		)
	`,
		run.ID,
		run.ProjectID,
		run.Branch,
		run.CommitSHA,
		run.Author,
		run.TriggerType,
		run.RunTimestamp,
		run.GinkgoVersion,
		run.SuiteDescription,
		run.SuitePath,
		run.TotalSpecs,
		run.PassedSpecs,
		run.FailedSpecs,
		run.SkippedSpecs,
		run.FlakedSpecs,
		run.PendingSpecs,
		run.Interrupted,
		run.TimedOut,
		run.DurationMS,
		run.Status,
		run.Environment,
		run.CreatedAt,
	)
	if err != nil {
		return domain.IntegrationTestRun{}, fmt.Errorf("insert integration test run: %w", err)
	}
	return run, nil
}

func (r *IntegrationTestRunRepository) GetLatestByProjectAndBranch(ctx context.Context, projectID string, branch string) (domain.IntegrationTestRun, error) {
	q := getQuerier(ctx, r.pool)
	var run domain.IntegrationTestRun
	err := q.QueryRow(ctx, `
		SELECT id, project_id, branch, commit_sha, COALESCE(author, ''), trigger_type, run_timestamp,
			COALESCE(ginkgo_version, ''), suite_description, suite_path, total_specs, passed_specs,
			failed_specs, skipped_specs, flaked_specs, pending_specs, interrupted, timed_out,
			duration_ms, status, environment, created_at
		FROM integration_test_runs
		WHERE project_id = $1 AND branch = $2
		ORDER BY run_timestamp DESC, created_at DESC
		LIMIT 1
	`, projectID, branch).Scan(
		&run.ID,
		&run.ProjectID,
		&run.Branch,
		&run.CommitSHA,
		&run.Author,
		&run.TriggerType,
		&run.RunTimestamp,
		&run.GinkgoVersion,
		&run.SuiteDescription,
		&run.SuitePath,
		&run.TotalSpecs,
		&run.PassedSpecs,
		&run.FailedSpecs,
		&run.SkippedSpecs,
		&run.FlakedSpecs,
		&run.PendingSpecs,
		&run.Interrupted,
		&run.TimedOut,
		&run.DurationMS,
		&run.Status,
		&run.Environment,
		&run.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.IntegrationTestRun{}, domain.ErrNotFound
		}
		return domain.IntegrationTestRun{}, fmt.Errorf("query latest integration run by project and branch: %w", err)
	}
	return run, nil
}

func (r *IntegrationTestRunRepository) GetLatestByProject(ctx context.Context, projectID string) (domain.IntegrationTestRun, error) {
	q := getQuerier(ctx, r.pool)
	var run domain.IntegrationTestRun
	err := q.QueryRow(ctx, `
		SELECT id, project_id, branch, commit_sha, COALESCE(author, ''), trigger_type, run_timestamp,
			COALESCE(ginkgo_version, ''), suite_description, suite_path, total_specs, passed_specs,
			failed_specs, skipped_specs, flaked_specs, pending_specs, interrupted, timed_out,
			duration_ms, status, environment, created_at
		FROM integration_test_runs
		WHERE project_id = $1
		ORDER BY run_timestamp DESC, created_at DESC
		LIMIT 1
	`, projectID).Scan(
		&run.ID,
		&run.ProjectID,
		&run.Branch,
		&run.CommitSHA,
		&run.Author,
		&run.TriggerType,
		&run.RunTimestamp,
		&run.GinkgoVersion,
		&run.SuiteDescription,
		&run.SuitePath,
		&run.TotalSpecs,
		&run.PassedSpecs,
		&run.FailedSpecs,
		&run.SkippedSpecs,
		&run.FlakedSpecs,
		&run.PendingSpecs,
		&run.Interrupted,
		&run.TimedOut,
		&run.DurationMS,
		&run.Status,
		&run.Environment,
		&run.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.IntegrationTestRun{}, domain.ErrNotFound
		}
		return domain.IntegrationTestRun{}, fmt.Errorf("query latest integration run by project: %w", err)
	}
	return run, nil
}

func (r *IntegrationTestRunRepository) GetByID(ctx context.Context, projectID string, runID string) (domain.IntegrationTestRun, error) {
	q := getQuerier(ctx, r.pool)
	var run domain.IntegrationTestRun
	err := q.QueryRow(ctx, `
		SELECT id, project_id, branch, commit_sha, COALESCE(author, ''), trigger_type, run_timestamp,
			COALESCE(ginkgo_version, ''), suite_description, suite_path, total_specs, passed_specs,
			failed_specs, skipped_specs, flaked_specs, pending_specs, interrupted, timed_out,
			duration_ms, status, environment, created_at
		FROM integration_test_runs
		WHERE project_id = $1 AND id = $2
		LIMIT 1
	`, projectID, runID).Scan(
		&run.ID,
		&run.ProjectID,
		&run.Branch,
		&run.CommitSHA,
		&run.Author,
		&run.TriggerType,
		&run.RunTimestamp,
		&run.GinkgoVersion,
		&run.SuiteDescription,
		&run.SuitePath,
		&run.TotalSpecs,
		&run.PassedSpecs,
		&run.FailedSpecs,
		&run.SkippedSpecs,
		&run.FlakedSpecs,
		&run.PendingSpecs,
		&run.Interrupted,
		&run.TimedOut,
		&run.DurationMS,
		&run.Status,
		&run.Environment,
		&run.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.IntegrationTestRun{}, domain.ErrNotFound
		}
		return domain.IntegrationTestRun{}, fmt.Errorf("query integration run by id: %w", err)
	}
	return run, nil
}

func (r *IntegrationTestRunRepository) ListByProject(ctx context.Context, projectID string, branch string, status string, environment string, from *time.Time, to *time.Time, page int, pageSize int) ([]domain.IntegrationTestRun, int, error) {
	q := getQuerier(ctx, r.pool)
	offset := (page - 1) * pageSize

	where := "WHERE project_id = $1"
	args := []any{projectID}
	idx := 2

	if branch != "" {
		where += fmt.Sprintf(" AND branch = $%d", idx)
		args = append(args, branch)
		idx++
	}
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, status)
		idx++
	}
	if environment != "" {
		if environment == "none" {
			where += " AND environment IS NULL"
		} else {
			where += fmt.Sprintf(" AND environment = $%d", idx)
			args = append(args, environment)
			idx++
		}
	}
	if from != nil {
		where += fmt.Sprintf(" AND run_timestamp >= $%d", idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		where += fmt.Sprintf(" AND run_timestamp <= $%d", idx)
		args = append(args, *to)
		idx++
	}

	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM integration_test_runs %s", where)
	var total int
	if err := q.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count integration runs: %w", err)
	}

	listSQL := fmt.Sprintf(`
		SELECT id, project_id, branch, commit_sha, COALESCE(author, ''), trigger_type, run_timestamp,
			COALESCE(ginkgo_version, ''), suite_description, suite_path, total_specs, passed_specs,
			failed_specs, skipped_specs, flaked_specs, pending_specs, interrupted, timed_out,
			duration_ms, status, environment, created_at
		FROM integration_test_runs
		%s
		ORDER BY run_timestamp DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, idx, idx+1)
	args = append(args, pageSize, offset)

	rows, err := q.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list integration runs: %w", err)
	}
	defer rows.Close()

	runs := make([]domain.IntegrationTestRun, 0)
	for rows.Next() {
		var run domain.IntegrationTestRun
		if err := rows.Scan(
			&run.ID,
			&run.ProjectID,
			&run.Branch,
			&run.CommitSHA,
			&run.Author,
			&run.TriggerType,
			&run.RunTimestamp,
			&run.GinkgoVersion,
			&run.SuiteDescription,
			&run.SuitePath,
			&run.TotalSpecs,
			&run.PassedSpecs,
			&run.FailedSpecs,
			&run.SkippedSpecs,
			&run.FlakedSpecs,
			&run.PendingSpecs,
			&run.Interrupted,
			&run.TimedOut,
			&run.DurationMS,
			&run.Status,
			&run.Environment,
			&run.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan integration run: %w", err)
		}
		runs = append(runs, run)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate integration run rows: %w", err)
	}

	return runs, total, nil
}

func (r *IntegrationTestRunRepository) HeatmapData(ctx context.Context, branch string, status string, runsPerProject int) ([]application.IntegrationHeatmapRow, error) {
	q := getQuerier(ctx, r.pool)

	where := "WHERE 1=1"
	args := []any{}
	idx := 1

	if branch != "" {
		where += fmt.Sprintf(" AND itr.branch = $%d", idx)
		args = append(args, branch)
		idx++
	}
	if status != "" {
		where += fmt.Sprintf(" AND itr.status = $%d", idx)
		args = append(args, status)
		idx++
	}

	args = append(args, runsPerProject)

	sql := fmt.Sprintf(`
		WITH ranked AS (
			SELECT
				itr.id              AS run_id,
				itr.project_id,
				itr.branch,
				itr.commit_sha,
				itr.run_timestamp,
				itr.passed_specs,
				itr.total_specs,
				itr.status,
				itr.environment,
				COALESCE(p.name, '')        AS project_name,
				p.project_key,
				COALESCE(p.group_name, '')  AS project_group,
				ROW_NUMBER() OVER (
					PARTITION BY itr.project_id
					ORDER BY itr.run_timestamp DESC, itr.created_at DESC
				) AS rn
			FROM integration_test_runs itr
			JOIN projects p ON p.id = itr.project_id
			%s
		)
		SELECT run_id, project_id, project_name, project_key, project_group,
		       branch, commit_sha, run_timestamp, passed_specs, total_specs, status, environment
		FROM ranked
		WHERE rn <= $%d
		ORDER BY
			CASE WHEN project_group = '' THEN 1 ELSE 0 END ASC,
			project_group ASC,
			project_name ASC,
			run_timestamp DESC
	`, where, idx)

	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query heatmap data: %w", err)
	}
	defer rows.Close()

	result := make([]application.IntegrationHeatmapRow, 0)
	for rows.Next() {
		var row application.IntegrationHeatmapRow
		if err := rows.Scan(
			&row.RunID,
			&row.ProjectID,
			&row.ProjectName,
			&row.ProjectKey,
			&row.ProjectGroup,
			&row.Branch,
			&row.CommitSHA,
			&row.RunTimestamp,
			&row.PassedSpecs,
			&row.TotalSpecs,
			&row.Status,
			&row.Environment,
		); err != nil {
			return nil, fmt.Errorf("scan heatmap row: %w", err)
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate heatmap rows: %w", err)
	}

	return result, nil
}
