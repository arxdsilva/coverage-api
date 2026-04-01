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

type CoverageRunRepository struct {
	pool *pgxpool.Pool
}

func NewCoverageRunRepository(pool *pgxpool.Pool) *CoverageRunRepository {
	return &CoverageRunRepository{pool: pool}
}

func (r *CoverageRunRepository) Create(ctx context.Context, run domain.CoverageRun) (domain.CoverageRun, error) {
	q := getQuerier(ctx, r.pool)
	_, err := q.Exec(ctx, `
		INSERT INTO coverage_runs (
			id, project_id, branch, commit_sha, author, trigger_type, run_timestamp, total_coverage_percent, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		run.ID,
		run.ProjectID,
		run.Branch,
		run.CommitSHA,
		run.Author,
		run.TriggerType,
		run.RunTimestamp,
		run.TotalCoveragePercent,
		run.CreatedAt,
	)
	if err != nil {
		return domain.CoverageRun{}, fmt.Errorf("insert coverage run: %w", err)
	}
	return run, nil
}

func (r *CoverageRunRepository) GetLatestByProjectAndBranch(ctx context.Context, projectID string, branch string) (domain.CoverageRun, error) {
	q := getQuerier(ctx, r.pool)
	var run domain.CoverageRun
	err := q.QueryRow(ctx, `
		SELECT id, project_id, branch, commit_sha, COALESCE(author, ''), trigger_type, run_timestamp, total_coverage_percent, created_at
		FROM coverage_runs
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
		&run.TotalCoveragePercent,
		&run.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CoverageRun{}, domain.ErrNotFound
		}
		return domain.CoverageRun{}, fmt.Errorf("query latest run by project and branch: %w", err)
	}

	return run, nil
}

func (r *CoverageRunRepository) GetLatestByProject(ctx context.Context, projectID string) (domain.CoverageRun, error) {
	q := getQuerier(ctx, r.pool)
	var run domain.CoverageRun
	err := q.QueryRow(ctx, `
		SELECT id, project_id, branch, commit_sha, COALESCE(author, ''), trigger_type, run_timestamp, total_coverage_percent, created_at
		FROM coverage_runs
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
		&run.TotalCoveragePercent,
		&run.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CoverageRun{}, domain.ErrNotFound
		}
		return domain.CoverageRun{}, fmt.Errorf("query latest run by project: %w", err)
	}

	return run, nil
}

func (r *CoverageRunRepository) ListByProject(
	ctx context.Context,
	projectID string,
	branch string,
	from *time.Time,
	to *time.Time,
	page int,
	pageSize int,
) ([]domain.CoverageRun, int, error) {
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

	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM coverage_runs %s", where)
	var total int
	if err := q.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count coverage runs: %w", err)
	}

	listSQL := fmt.Sprintf(`
		SELECT id, project_id, branch, commit_sha, COALESCE(author, ''), trigger_type, run_timestamp, total_coverage_percent, created_at
		FROM coverage_runs
		%s
		ORDER BY run_timestamp DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, idx, idx+1)
	args = append(args, pageSize, offset)

	rows, err := q.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list coverage runs: %w", err)
	}
	defer rows.Close()

	runs := make([]domain.CoverageRun, 0)
	for rows.Next() {
		var run domain.CoverageRun
		if err := rows.Scan(
			&run.ID,
			&run.ProjectID,
			&run.Branch,
			&run.CommitSHA,
			&run.Author,
			&run.TriggerType,
			&run.RunTimestamp,
			&run.TotalCoveragePercent,
			&run.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan coverage run: %w", err)
		}
		runs = append(runs, run)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate coverage run rows: %w", err)
	}

	return runs, total, nil
}

func (r *CoverageRunRepository) ListBranchesByProject(ctx context.Context, projectID string) ([]string, error) {
	q := getQuerier(ctx, r.pool)
	rows, err := q.Query(ctx, `
		SELECT DISTINCT branch FROM coverage_runs
		WHERE project_id = $1
		ORDER BY branch ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query branches: %w", err)
	}
	defer rows.Close()

	var branches []string
	for rows.Next() {
		var branch string
		if err := rows.Scan(&branch); err != nil {
			return nil, fmt.Errorf("scan branch: %w", err)
		}
		branches = append(branches, branch)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate branch rows: %w", err)
	}

	return branches, nil
}

func (r *CoverageRunRepository) ListContributorsByProjectAndBranch(ctx context.Context, projectID string, branch string, limit int) ([]application.ContributorSummary, error) {
	q := getQuerier(ctx, r.pool)
	if limit <= 0 {
		limit = 10
	}

	rows, err := q.Query(ctx, `
		SELECT
			COALESCE(NULLIF(author, ''), 'unknown') AS author,
			COUNT(DISTINCT commit_sha) AS commit_count,
			COUNT(*) AS run_count,
			AVG(total_coverage_percent)::float8 AS average_coverage_percent,
			MAX(run_timestamp) AS latest_run_timestamp
		FROM coverage_runs
		WHERE project_id = $1 AND branch = $2
		GROUP BY COALESCE(NULLIF(author, ''), 'unknown')
		ORDER BY commit_count DESC, run_count DESC, latest_run_timestamp DESC, author ASC
		LIMIT $3
	`, projectID, branch, limit)
	if err != nil {
		return nil, fmt.Errorf("query contributors: %w", err)
	}
	defer rows.Close()

	contributors := make([]application.ContributorSummary, 0)
	for rows.Next() {
		var contributor application.ContributorSummary
		if err := rows.Scan(
			&contributor.Author,
			&contributor.CommitCount,
			&contributor.RunCount,
			&contributor.AverageCoveragePercent,
			&contributor.LatestRunTimestamp,
		); err != nil {
			return nil, fmt.Errorf("scan contributor: %w", err)
		}
		contributors = append(contributors, contributor)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate contributor rows: %w", err)
	}

	return contributors, nil
}
