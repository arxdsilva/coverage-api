package application

import (
	"context"
	"errors"
	"time"

	"github.com/arxdsilva/opencoverage/internal/domain"
)

type ContributorSummary struct {
	Author                 string    `json:"author"`
	CommitCount            int       `json:"commitCount"`
	RunCount               int       `json:"runCount"`
	AverageCoveragePercent float64   `json:"averageCoveragePercent"`
	LatestRunTimestamp     time.Time `json:"latestRunTimestamp"`
}

type ContributorResponse struct {
	Author                 string  `json:"author"`
	CommitCount            int     `json:"commitCount"`
	RunCount               int     `json:"runCount"`
	AverageCoveragePercent float64 `json:"averageCoveragePercent"`
	LatestRunTimestamp     string  `json:"latestRunTimestamp"`
}

type ListContributorsInput struct {
	ProjectID string
	Limit     int
}

type ListContributorsOutput struct {
	ProjectID     string                `json:"projectId"`
	DefaultBranch string                `json:"defaultBranch"`
	Contributors  []ContributorResponse `json:"contributors"`
}

type ListContributorsUseCase struct {
	projects ProjectRepository
	runs     CoverageRunRepository
}

func NewListContributorsUseCase(projects ProjectRepository, runs CoverageRunRepository) *ListContributorsUseCase {
	return &ListContributorsUseCase{projects: projects, runs: runs}
}

func (uc *ListContributorsUseCase) Execute(ctx context.Context, in ListContributorsInput) (ListContributorsOutput, error) {
	project, err := uc.projects.GetByID(ctx, in.ProjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ListContributorsOutput{}, NewNotFound("project not found", map[string]any{"projectId": in.ProjectID})
		}
		return ListContributorsOutput{}, NewInternal("failed to load project", err)
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 25 {
		limit = 25
	}

	contributors, err := uc.runs.ListContributorsByProjectAndBranch(ctx, in.ProjectID, project.DefaultBranch, limit)
	if err != nil {
		return ListContributorsOutput{}, NewInternal("failed to list contributors", err)
	}

	out := make([]ContributorResponse, 0, len(contributors))
	for _, contributor := range contributors {
		out = append(out, ContributorResponse{
			Author:                 contributor.Author,
			CommitCount:            contributor.CommitCount,
			RunCount:               contributor.RunCount,
			AverageCoveragePercent: contributor.AverageCoveragePercent,
			LatestRunTimestamp:     contributor.LatestRunTimestamp.UTC().Format(time.RFC3339),
		})
	}

	return ListContributorsOutput{
		ProjectID:     in.ProjectID,
		DefaultBranch: project.DefaultBranch,
		Contributors:  out,
	}, nil
}
