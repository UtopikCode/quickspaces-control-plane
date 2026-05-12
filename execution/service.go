package execution

import (
	"context"

	"github.com/UtopikCode/quickspaces-control-plane/domain"
	"github.com/UtopikCode/quickspaces-execution-contracts"
)

type ExecutionService struct {
	adapter contracts.ExecutionAdapter
}

func NewExecutionService(adapter contracts.ExecutionAdapter) *ExecutionService {
	return &ExecutionService{adapter: adapter}
}

func (s *ExecutionService) StartWorkspace(ctx context.Context, workspace *domain.Workspace) error {
	return s.adapter.StartWorkspace(ctx, toWorkspaceSpec(workspace))
}

func (s *ExecutionService) StopWorkspace(ctx context.Context, workspace *domain.Workspace) error {
	return s.adapter.StopWorkspace(ctx, toWorkspaceSpec(workspace))
}

func (s *ExecutionService) GetWorkspaceStatus(ctx context.Context, workspace *domain.Workspace) (string, error) {
	return s.adapter.GetWorkspaceStatus(ctx, toWorkspaceSpec(workspace))
}

func toWorkspaceSpec(workspace *domain.Workspace) contracts.WorkspaceSpec {
	return contracts.WorkspaceSpec{
		ID:               workspace.ID,
		Repo:             workspace.Repo,
		Owner:            workspace.Owner,
		Ref:              workspace.Ref,
		DesiredState:     workspace.DesiredState,
		ExecutionProfile: workspace.ExecutionProfile,
	}
}
