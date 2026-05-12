package adapters

import (
	"context"

	"github.com/UtopikCode/quickspaces-execution-contracts"
)

type AWSExecutionAdapter struct{}

func NewAWSExecutionAdapter() contracts.ExecutionAdapter {
	return &AWSExecutionAdapter{}
}

func (a *AWSExecutionAdapter) StartWorkspace(ctx context.Context, workspace contracts.WorkspaceSpec) error {
	// Control Plane is stateless and must delegate execution. This adapter does not implement AWS APIs.
	return nil
}

func (a *AWSExecutionAdapter) StopWorkspace(ctx context.Context, workspace contracts.WorkspaceSpec) error {
	return nil
}

func (a *AWSExecutionAdapter) GetWorkspaceStatus(ctx context.Context, workspace contracts.WorkspaceSpec) (string, error) {
	if workspace.DesiredState == "running" {
		return "running", nil
	}
	return "stopped", nil
}
