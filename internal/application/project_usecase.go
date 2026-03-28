package application

import (
	"context"
	"errors"

	"github.com/arxdsilva/coverage-api/internal/domain"
)

type GetProjectUseCase struct {
	projects ProjectRepository
}

func NewGetProjectUseCase(projects ProjectRepository) *GetProjectUseCase {
	return &GetProjectUseCase{projects: projects}
}

func (uc *GetProjectUseCase) Execute(ctx context.Context, projectID string) (ProjectResponse, error) {
	project, err := uc.projects.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ProjectResponse{}, NewNotFound("project not found", map[string]any{"projectId": projectID})
		}
		return ProjectResponse{}, NewInternal("failed to load project", err)
	}

	return ProjectResponse{
		ID:                     project.ID,
		ProjectKey:             project.ProjectKey,
		Name:                   project.Name,
		DefaultBranch:          project.DefaultBranch,
		GlobalThresholdPercent: project.GlobalThresholdPercent,
		Created:                false,
	}, nil
}
