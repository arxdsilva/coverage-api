package application

import (
	"context"
	"time"

	"github.com/arxdsilva/opencoverage/internal/domain"
)

type ProjectRepository interface {
	GetByKey(ctx context.Context, projectKey string) (domain.Project, error)
	GetByID(ctx context.Context, projectID string) (domain.Project, error)
	List(ctx context.Context, page int, pageSize int) ([]domain.Project, int, error)
	Create(ctx context.Context, project domain.Project) (domain.Project, error)
}

type CoverageRunRepository interface {
	Create(ctx context.Context, run domain.CoverageRun) (domain.CoverageRun, error)
	GetLatestByProjectAndBranch(ctx context.Context, projectID string, branch string) (domain.CoverageRun, error)
	GetLatestByProject(ctx context.Context, projectID string) (domain.CoverageRun, error)
	ListByProject(ctx context.Context, projectID string, branch string, from *time.Time, to *time.Time, page int, pageSize int) ([]domain.CoverageRun, int, error)
	ListBranchesByProject(ctx context.Context, projectID string) ([]string, error)
	ListContributorsByProjectAndBranch(ctx context.Context, projectID string, branch string, limit int) ([]ContributorSummary, error)
}

type PackageCoverageRepository interface {
	CreateBatch(ctx context.Context, packages []domain.PackageCoverage) error
	ListByRunID(ctx context.Context, runID string) ([]domain.PackageCoverage, error)
}

type IntegrationHeatmapRow struct {
	RunID        string
	ProjectID    string
	ProjectName  string
	ProjectKey   string
	ProjectGroup string // empty string means no group
	Branch       string
	CommitSHA    string
	RunTimestamp time.Time
	PassedSpecs  int
	TotalSpecs   int
	Status       string
	Environment  *string
}

type IntegrationTestRunRepository interface {
	Create(ctx context.Context, run domain.IntegrationTestRun) (domain.IntegrationTestRun, error)
	GetLatestByProjectAndBranch(ctx context.Context, projectID string, branch string) (domain.IntegrationTestRun, error)
	GetLatestByProject(ctx context.Context, projectID string) (domain.IntegrationTestRun, error)
	GetByID(ctx context.Context, projectID string, runID string) (domain.IntegrationTestRun, error)
	ListByProject(ctx context.Context, projectID string, branch string, status string, environment string, from *time.Time, to *time.Time, page int, pageSize int) ([]domain.IntegrationTestRun, int, error)
	HeatmapData(ctx context.Context, branch string, status string, runsPerProject int) ([]IntegrationHeatmapRow, error)
}

type IntegrationSpecResultRepository interface {
	CreateBatch(ctx context.Context, specs []domain.IntegrationSpecResult) error
	ListByRunID(ctx context.Context, runID string) ([]domain.IntegrationSpecResult, error)
	ListFailedByRunID(ctx context.Context, runID string) ([]domain.IntegrationSpecResult, error)
}

type APIKeyAuthenticator interface {
	Authenticate(ctx context.Context, apiKey string) error
	WantedAPIKey() string
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() string
}

type TransactionManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
