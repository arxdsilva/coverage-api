package postgres

import (
	"context"
	"fmt"

	"github.com/arxdsilva/opencoverage/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IntegrationSpecResultRepository struct {
	pool *pgxpool.Pool
}

func NewIntegrationSpecResultRepository(pool *pgxpool.Pool) *IntegrationSpecResultRepository {
	return &IntegrationSpecResultRepository{pool: pool}
}

func (r *IntegrationSpecResultRepository) CreateBatch(ctx context.Context, specs []domain.IntegrationSpecResult) error {
	if len(specs) == 0 {
		return nil
	}

	q := getQuerier(ctx, r.pool)
	for _, spec := range specs {
		_, err := q.Exec(ctx, `
			INSERT INTO integration_spec_results (
				id, integration_run_id, spec_path, leaf_node_text, state, duration_ms,
				failure_message, failure_location_file, failure_location_line
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
			spec.ID,
			spec.IntegrationRunID,
			spec.SpecPath,
			spec.LeafNodeText,
			spec.State,
			spec.DurationMS,
			spec.FailureMessage,
			spec.FailureLocationFile,
			spec.FailureLocationLine,
		)
		if err != nil {
			return fmt.Errorf("insert integration spec result: %w", err)
		}
	}

	return nil
}

func (r *IntegrationSpecResultRepository) ListByRunID(ctx context.Context, runID string) ([]domain.IntegrationSpecResult, error) {
	q := getQuerier(ctx, r.pool)
	rows, err := q.Query(ctx, `
		SELECT id, integration_run_id, spec_path, leaf_node_text, state, duration_ms,
			failure_message, failure_location_file, failure_location_line
		FROM integration_spec_results
		WHERE integration_run_id = $1
		ORDER BY spec_path ASC
	`, runID)
	if err != nil {
		return nil, fmt.Errorf("query integration spec results: %w", err)
	}
	defer rows.Close()

	specs := make([]domain.IntegrationSpecResult, 0)
	for rows.Next() {
		var spec domain.IntegrationSpecResult
		if err := rows.Scan(
			&spec.ID,
			&spec.IntegrationRunID,
			&spec.SpecPath,
			&spec.LeafNodeText,
			&spec.State,
			&spec.DurationMS,
			&spec.FailureMessage,
			&spec.FailureLocationFile,
			&spec.FailureLocationLine,
		); err != nil {
			return nil, fmt.Errorf("scan integration spec result: %w", err)
		}
		specs = append(specs, spec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate integration spec rows: %w", err)
	}

	return specs, nil
}

func (r *IntegrationSpecResultRepository) ListFailedByRunID(ctx context.Context, runID string) ([]domain.IntegrationSpecResult, error) {
	q := getQuerier(ctx, r.pool)
	rows, err := q.Query(ctx, `
		SELECT id, integration_run_id, spec_path, leaf_node_text, state, duration_ms,
			failure_message, failure_location_file, failure_location_line
		FROM integration_spec_results
		WHERE integration_run_id = $1 AND state IN ('failed', 'flaky')
		ORDER BY spec_path ASC
	`, runID)
	if err != nil {
		return nil, fmt.Errorf("query failed integration spec results: %w", err)
	}
	defer rows.Close()

	specs := make([]domain.IntegrationSpecResult, 0)
	for rows.Next() {
		var spec domain.IntegrationSpecResult
		if err := rows.Scan(
			&spec.ID,
			&spec.IntegrationRunID,
			&spec.SpecPath,
			&spec.LeafNodeText,
			&spec.State,
			&spec.DurationMS,
			&spec.FailureMessage,
			&spec.FailureLocationFile,
			&spec.FailureLocationLine,
		); err != nil {
			return nil, fmt.Errorf("scan failed integration spec result: %w", err)
		}
		specs = append(specs, spec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate failed integration spec rows: %w", err)
	}

	return specs, nil
}
