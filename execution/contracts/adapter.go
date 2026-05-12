package contracts

import (
	"context"
	"encoding/json"
)

type ExecutionProfile = json.RawMessage

type WorkspaceSpec struct {
	ID               string           `json:"id"`
	Repo             string           `json:"repo"`
	Owner            string           `json:"owner"`
	Ref              string           `json:"ref"`
	DesiredState     string           `json:"desiredState"`
	ExecutionProfile ExecutionProfile `json:"executionProfile"`
}

type ExecutionAdapter interface {
	StartWorkspace(ctx context.Context, workspace WorkspaceSpec) error
	StopWorkspace(ctx context.Context, workspace WorkspaceSpec) error
	GetWorkspaceStatus(ctx context.Context, workspace WorkspaceSpec) (string, error)
}
