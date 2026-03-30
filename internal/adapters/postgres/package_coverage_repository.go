package postgres

import (
	"context"
	"fmt"

	"github.com/arxdsilva/opencoverage/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PackageCoverageRepository struct {
	pool *pgxpool.Pool
}

func NewPackageCoverageRepository(pool *pgxpool.Pool) *PackageCoverageRepository {
	return &PackageCoverageRepository{pool: pool}
}

func (r *PackageCoverageRepository) CreateBatch(ctx context.Context, packages []domain.PackageCoverage) error {
	q := getQuerier(ctx, r.pool)
	for _, p := range packages {
		_, err := q.Exec(ctx, `
			INSERT INTO package_coverages (id, run_id, package_import_path, coverage_percent)
			VALUES ($1, $2, $3, $4)
		`, p.ID, p.RunID, p.PackageImportPath, p.CoveragePercent)
		if err != nil {
			return fmt.Errorf("insert package coverage: %w", err)
		}
	}
	return nil
}

func (r *PackageCoverageRepository) ListByRunID(ctx context.Context, runID string) ([]domain.PackageCoverage, error) {
	q := getQuerier(ctx, r.pool)
	rows, err := q.Query(ctx, `
		SELECT id, run_id, package_import_path, coverage_percent
		FROM package_coverages
		WHERE run_id = $1
		ORDER BY package_import_path ASC
	`, runID)
	if err != nil {
		return nil, fmt.Errorf("query package coverages by run id: %w", err)
	}
	defer rows.Close()

	out := make([]domain.PackageCoverage, 0)
	for rows.Next() {
		var p domain.PackageCoverage
		if err := rows.Scan(&p.ID, &p.RunID, &p.PackageImportPath, &p.CoveragePercent); err != nil {
			return nil, fmt.Errorf("scan package coverage: %w", err)
		}
		out = append(out, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate package coverage rows: %w", err)
	}

	return out, nil
}
