package adapters

import (
	"context"

	"github.com/UtopikCode/quickspaces-execution-contracts"
)

type LocalExecutionAdapter struct{}

func NewLocalExecutionAdapter() contracts.ExecutionAdapter {
	return &LocalExecutionAdapter{}
}

func (l *LocalExecutionAdapter) StartWorkspace(ctx context.Context, workspace contracts.WorkspaceSpec) error {
	return nil
}

func (l *LocalExecutionAdapter) StopWorkspace(ctx context.Context, workspace contracts.WorkspaceSpec) error {
	return nil
}

func (l *LocalExecutionAdapter) GetWorkspaceStatus(ctx context.Context, workspace contracts.WorkspaceSpec) (string, error) {
	if workspace.DesiredState == "running" {
		return "running", nil
	}
	return "stopped", nil
}
