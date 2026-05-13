package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arxdsilva/opencoverage/internal/domain"
)

func validIngestInput() IngestCoverageRunInput {
	return IngestCoverageRunInput{
		ProjectKey:           "acme/coverage",
		ProjectName:          "coverage",
		DefaultBranch:        "main",
		Branch:               "main",
		CommitSHA:            "abc123",
		Author:               "alice",
		TriggerType:          "push",
		RunTimestamp:         "2026-01-01T00:00:00Z",
		TotalCoveragePercent: 80.0,
		Packages: []IngestPackageInput{
			{ImportPath: "github.com/foo/bar", CoveragePercent: 80.0},
		},
	}
}

func newIngestCoverageRunUseCase(
	projects ProjectRepository,
	runs CoverageRunRepository,
	packages PackageCoverageRepository,
) *IngestCoverageRunUseCase {
	return NewIngestCoverageRunUseCase(
		projects,
		runs,
		packages,
		&stubTransactionManager{},
		&stubIDGenerator{},
		&stubClock{},
	)
}


func TestIngestCoverageRunExecute(t *testing.T) {
	t.Run("invalid runTimestamp returns error", func(t *testing.T) {
		in := validIngestInput()
		in.RunTimestamp = "not-a-date"

		uc := newIngestCoverageRunUseCase(
			&stubProjectRepository{},
			&stubCoverageRunRepository{},
			&stubPackageCoverageRepository{},
		)
		_, err := uc.Execute(context.Background(), in)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("new project is created when it does not exist", func(t *testing.T) {
		project := &stubProjectRepository{existing: nil}
		runs := &stubCoverageRunRepository{}

		uc := newIngestCoverageRunUseCase(project, runs, &stubPackageCoverageRepository{})

		out, err := uc.Execute(context.Background(), validIngestInput())

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !out.Project.Created {
			t.Errorf("expected project.Created = true for a new project")
		}

		if out.Project.ProjectKey != "acme/coverage" {
			t.Errorf("expected project key acme/coverage, got %s", out.Project.ProjectKey)
		}
	})

	t.Run("existing project is not re-created", func(t *testing.T) {
		// Arrange: existing is set → GetByKey returns it
		existing := domain.Project{
			ID:                     "proj-1",
			ProjectKey:             "acme/coverage",
			DefaultBranch:          "main",
			GlobalThresholdPercent: 70.0,
		}
		project := &stubProjectRepository{existing: &existing}
		runs := &stubCoverageRunRepository{}

		uc := newIngestCoverageRunUseCase(project, runs, &stubPackageCoverageRepository{})

		out, err := uc.Execute(context.Background(), validIngestInput())

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if out.Project.Created {
			t.Errorf("expected project.Created = false for an existing project")
		}
	})

	t.Run("threshold is updated when project exists and threshold is provided", func(t *testing.T) {
		newThreshold := 90.0
		existing := domain.Project{
			ID:                     "proj-1",
			ProjectKey:             "acme/coverage",
			DefaultBranch:          "main",
			GlobalThresholdPercent: 70.0,
		}
		afterUpdate := domain.Project{
			ID:                     "proj-1",
			ProjectKey:             "acme/coverage",
			DefaultBranch:          "main",
			GlobalThresholdPercent: newThreshold,
		}
		project := &stubProjectRepository{existing: &existing, project: afterUpdate}
		runs := &stubCoverageRunRepository{}

		in := validIngestInput()
		in.ThresholdPercent = &newThreshold

		uc := newIngestCoverageRunUseCase(project, runs, &stubPackageCoverageRepository{})

		out, err := uc.Execute(context.Background(), in)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if out.Project.GlobalThresholdPercent != 90.0 {
			t.Errorf("expected threshold 90.0, got %f", out.Project.GlobalThresholdPercent)
		}
	})

	t.Run("no baseline run means nil previous coverage in comparison", func(t *testing.T) {
		project := &stubProjectRepository{existing: nil}
		runs := &stubCoverageRunRepository{baseline: nil}

		uc := newIngestCoverageRunUseCase(project, runs, &stubPackageCoverageRepository{})

		out, err := uc.Execute(context.Background(), validIngestInput())

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if out.Comparison.PreviousTotalCoveragePercent != nil {
			t.Errorf("expected nil previous coverage, got %v", out.Comparison.PreviousTotalCoveragePercent)
		}
	})

	t.Run("baseline run is used in comparison", func(t *testing.T) {
		baselineCoverage := 60.0
		baseline := domain.CoverageRun{
			ID:                   "run-baseline",
			TotalCoveragePercent: baselineCoverage,
		}
		newThreshold := 90.0
		existing := domain.Project{
			ID:                     "proj-1",
			ProjectKey:             "acme/coverage",
			DefaultBranch:          "main",
			GlobalThresholdPercent: 70.0,
		}
		afterUpdate := domain.Project{
			ID:                     "proj-1",
			ProjectKey:             "acme/coverage",
			DefaultBranch:          "main",
			GlobalThresholdPercent: newThreshold,
		}
		project := &stubProjectRepository{existing: &existing, project: afterUpdate}
		runs := &stubCoverageRunRepository{baseline: &baseline}
		in := validIngestInput()
		in.TotalCoveragePercent = 90.0

		uc := newIngestCoverageRunUseCase(project, runs, &stubPackageCoverageRepository{})

		out, err := uc.Execute(context.Background(), in)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if out.Comparison.PreviousTotalCoveragePercent == nil {
			t.Fatal("expected previous coverage to be set")
		}
		if *out.Comparison.PreviousTotalCoveragePercent != 60.0 {
			t.Errorf("expected previous coverage 60.0, got %f", *out.Comparison.PreviousTotalCoveragePercent)
		}
	})
}

func TestResolveOrCreateProject(t *testing.T) {
	t.Run("creates new project when not found", func(t *testing.T) {
		projectRepo := &stubProjectRepository{existing: nil}
		run := &stubCoverageRunRepository{}

		uc := newIngestCoverageRunUseCase(projectRepo, run, &stubPackageCoverageRepository{})
		project, created, err := uc.resolveOrCreateProject(context.Background(), validIngestInput())

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !created {
			t.Error("expected project to be created")
		}
		if project.ProjectKey != "acme/coverage" {
			t.Errorf("expected project key acme/coverage, got %s", project.ProjectKey)
		}
	})
	t.Run("returns existing project when found", func(t *testing.T) {
		existing := domain.Project{
			ID:                     "proj-1",
			ProjectKey:             "acme/coverage",
			DefaultBranch:          "main",
			GlobalThresholdPercent: 70.0,
		}
		projectRepo := &stubProjectRepository{existing: &existing}
		run := &stubCoverageRunRepository{}
		uc := newIngestCoverageRunUseCase(projectRepo, run, &stubPackageCoverageRepository{})
		project, created, err := uc.resolveOrCreateProject(context.Background(), validIngestInput())

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if created {
			t.Error("expected project to be existing")
		}
		if project.ProjectKey != "acme/coverage" {
			t.Errorf("expected project key acme/coverage, got %s", project.ProjectKey)
		}
	})
	t.Run("returns error when Create returns error", func(t *testing.T) {
		projectRepo := &stubProjectRepository{existing: nil, err: fmt.Errorf("failed to create project"), project: domain.Project{}}
		run := &stubCoverageRunRepository{}
		uc := newIngestCoverageRunUseCase(projectRepo, run, &stubPackageCoverageRepository{})
		_, _, err := uc.resolveOrCreateProject(context.Background(), validIngestInput())

		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestValidateIngestInput(t *testing.T) {
	test := []struct {
		name      string
		mutate    func(*IngestCoverageRunInput)
		wantErr   bool
		wantField string
	}{
		{
			name:    "valid input passes",
			mutate:  func(in *IngestCoverageRunInput) {},
			wantErr: false,
		},
		{
			name:      "missing projectKey returns error",
			mutate:    func(in *IngestCoverageRunInput) { in.ProjectKey = "" },
			wantErr:   true,
			wantField: "projectKey",
		},
		{
			name:      "missing branch returns error",
			mutate:    func(in *IngestCoverageRunInput) { in.Branch = "" },
			wantErr:   true,
			wantField: "branch",
		},
		{
			name:      "missing commitSha returns error",
			mutate:    func(in *IngestCoverageRunInput) { in.CommitSHA = "" },
			wantErr:   true,
			wantField: "commitSha",
		},
		{
			name:      "invalid triggerType returns error",
			mutate:    func(in *IngestCoverageRunInput) { in.TriggerType = "unknown" },
			wantErr:   true,
			wantField: "triggerType",
		},
		{
			name:      "valid triggerType push passes",
			mutate:    func(in *IngestCoverageRunInput) { in.TriggerType = "push" },
			wantErr:   false,
			wantField: "triggerType",
		},
		{
			name:      "valid triggerType pr passes",
			mutate:    func(in *IngestCoverageRunInput) { in.TriggerType = "pr" },
			wantErr:   false,
			wantField: "triggerType",
		},
		{
			name:      "valid triggerType manual passes",
			mutate:    func(in *IngestCoverageRunInput) { in.TriggerType = "manual" },
			wantErr:   false,
			wantField: "triggerType",
		},
		{
			name:      "totalCoveragePercent above 100 returns error",
			mutate:    func(in *IngestCoverageRunInput) { in.TotalCoveragePercent = 101.0 },
			wantErr:   true,
			wantField: "totalCoveragePercent",
		},
		{
			name:      "totalCoveragePercent below 0 returns error",
			mutate:    func(in *IngestCoverageRunInput) { in.TotalCoveragePercent = -1.0 },
			wantErr:   true,
			wantField: "totalCoveragePercent",
		},
		{
			name:      "totalCoveragePercent between 0 and 100 passes",
			mutate:    func(in *IngestCoverageRunInput) { in.TotalCoveragePercent = 50.0 },
			wantErr:   false,
			wantField: "totalCoveragePercent",
		},
		{
			name:      "thresholdPercent above 100 returns error",
			mutate:    func(in *IngestCoverageRunInput) { in.ThresholdPercent = func() *float64 { v := 101.0; return &v }() },
			wantErr:   true,
			wantField: "thresholdPercent",
		},
		{
			name:      "thresholdPercent below 0 returns error",
			mutate:    func(in *IngestCoverageRunInput) { in.ThresholdPercent = func() *float64 { v := -1.0; return &v }() },
			wantErr:   true,
			wantField: "thresholdPercent",
		},
		{
			name:      "valid thresholdPercent passes",
			mutate:    func(in *IngestCoverageRunInput) { in.ThresholdPercent = func() *float64 { v := 85.0; return &v }() },
			wantErr:   false,
			wantField: "thresholdPercent",
		},
		{
			name:      "empty packages returns error",
			mutate:    func(in *IngestCoverageRunInput) { in.Packages = nil },
			wantErr:   true,
			wantField: "packages",
		},
		{
			name: "package with empty importPath returns error",
			mutate: func(in *IngestCoverageRunInput) {
				in.Packages = []IngestPackageInput{{ImportPath: "", CoveragePercent: 80.0}}
			},
			wantErr:   true,
			wantField: "packages[0].importPath",
		},
		{
			name: "duplicate package importPath returns error",
			mutate: func(in *IngestCoverageRunInput) {
				in.Packages = []IngestPackageInput{
					{ImportPath: "github.com/foo/bar", CoveragePercent: 80.0},
					{ImportPath: "github.com/foo/bar", CoveragePercent: 70.0},
				}
			},
			wantErr: true,
		},
		{
			name: "package coveragePercent above 100 returns error",
			mutate: func(in *IngestCoverageRunInput) {
				in.Packages = []IngestPackageInput{{ImportPath: "github.com/foo/bar", CoveragePercent: 101.0}}
			},
			wantErr:   true,
			wantField: "packages[0].coveragePercent",
		},
		{
			name: "package coveragePercent below 0 returns error",
			mutate: func(in *IngestCoverageRunInput) {
				in.Packages = []IngestPackageInput{{ImportPath: "github.com/foo/bar", CoveragePercent: -1.0}}
			},
			wantErr:   true,
			wantField: "packages[0].coveragePercent",
		},
		{
			name: "package coveragePercent between 0 and 100 passes",
			mutate: func(in *IngestCoverageRunInput) {
				in.Packages = []IngestPackageInput{{ImportPath: "github.com/foo/bar", CoveragePercent: 50.0}}
			},
			wantErr:   false,
			wantField: "packages[0].coveragePercent",
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			in := validIngestInput()
			tt.mutate(&in)
			err := validateIngestInput(in)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.wantErr && tt.wantField != "" {
				appErr, ok := err.(*AppError)
				if !ok {
					t.Fatalf("expected *AppError, got %T", err)
				}
				field, _ := appErr.Details["field"].(string)
				if field != tt.wantField {
					t.Errorf("expected field=%q, got %q", tt.wantField, field)
				}
			}
		})
	}
}

// Stubs
type stubTransactionManager struct{}

func (s *stubTransactionManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type stubIDGenerator struct{ n int }

func (s *stubIDGenerator) NewID() string {
	s.n++
	return fmt.Sprintf("generated-id-%d", s.n)
}

type stubClock struct{}

func (s *stubClock) Now() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
}

type stubPackageCoverageRepository struct {
	createdPackages []domain.PackageCoverage
}

func (s *stubPackageCoverageRepository) CreateBatch(ctx context.Context, packages []domain.PackageCoverage) error {
	return nil
}

func (s *stubPackageCoverageRepository) ListByRunID(ctx context.Context, runID string) ([]domain.PackageCoverage, error) {
	return s.createdPackages, nil
}
