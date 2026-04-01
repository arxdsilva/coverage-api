package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/arxdsilva/opencoverage/internal/domain"
)

func TestListContributorsUseCaseExecute(t *testing.T) {
	t.Run("uses default branch and formats output", func(t *testing.T) {
		projectRepo := &stubProjectRepository{
			project: domain.Project{ID: "project-1", DefaultBranch: "main"},
		}
		runRepo := &stubCoverageRunRepository{
			contributors: []ContributorSummary{{
				Author:                 "alice",
				CommitCount:            4,
				RunCount:               5,
				AverageCoveragePercent: 81.25,
				LatestRunTimestamp:     time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
			}},
		}

		uc := NewListContributorsUseCase(projectRepo, runRepo)
		out, err := uc.Execute(context.Background(), ListContributorsInput{ProjectID: "project-1", Limit: 50})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if runRepo.branch != "main" {
			t.Fatalf("expected default branch main, got %s", runRepo.branch)
		}
		if runRepo.limit != 25 {
			t.Fatalf("expected capped limit 25, got %d", runRepo.limit)
		}
		if len(out.Contributors) != 1 {
			t.Fatalf("expected 1 contributor, got %d", len(out.Contributors))
		}
		if out.Contributors[0].LatestRunTimestamp != "2026-04-01T12:00:00Z" {
			t.Fatalf("unexpected timestamp %s", out.Contributors[0].LatestRunTimestamp)
		}
	})

	t.Run("returns not found when project is missing", func(t *testing.T) {
		projectRepo := &stubProjectRepository{err: domain.ErrNotFound}
		runRepo := &stubCoverageRunRepository{}

		uc := NewListContributorsUseCase(projectRepo, runRepo)
		_, err := uc.Execute(context.Background(), ListContributorsInput{ProjectID: "missing"})
		if err == nil {
			t.Fatalf("expected error")
		}

		var appErr *AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("expected AppError, got %T", err)
		}
		if appErr.Code != CodeNotFound {
			t.Fatalf("expected not_found, got %s", appErr.Code)
		}
	})
}

type stubProjectRepository struct {
	project domain.Project
	err    error
}

func (s *stubProjectRepository) GetByKey(ctx context.Context, projectKey string) (domain.Project, error) {
	panic("unexpected call")
}

func (s *stubProjectRepository) GetByID(ctx context.Context, projectID string) (domain.Project, error) {
	if s.err != nil {
		return domain.Project{}, s.err
	}
	return s.project, nil
}

func (s *stubProjectRepository) List(ctx context.Context, page int, pageSize int) ([]domain.Project, int, error) {
	panic("unexpected call")
}

func (s *stubProjectRepository) Create(ctx context.Context, project domain.Project) (domain.Project, error) {
	panic("unexpected call")
}

type stubCoverageRunRepository struct {
	contributors []ContributorSummary
	err          error
	branch       string
	limit        int
}

func (s *stubCoverageRunRepository) Create(ctx context.Context, run domain.CoverageRun) (domain.CoverageRun, error) {
	panic("unexpected call")
}

func (s *stubCoverageRunRepository) GetLatestByProjectAndBranch(ctx context.Context, projectID string, branch string) (domain.CoverageRun, error) {
	panic("unexpected call")
}

func (s *stubCoverageRunRepository) GetLatestByProject(ctx context.Context, projectID string) (domain.CoverageRun, error) {
	panic("unexpected call")
}

func (s *stubCoverageRunRepository) ListByProject(ctx context.Context, projectID string, branch string, from *time.Time, to *time.Time, page int, pageSize int) ([]domain.CoverageRun, int, error) {
	panic("unexpected call")
}

func (s *stubCoverageRunRepository) ListBranchesByProject(ctx context.Context, projectID string) ([]string, error) {
	panic("unexpected call")
}

func (s *stubCoverageRunRepository) ListContributorsByProjectAndBranch(ctx context.Context, projectID string, branch string, limit int) ([]ContributorSummary, error) {
	s.branch = branch
	s.limit = limit
	if s.err != nil {
		return nil, s.err
	}
	return s.contributors, nil
}