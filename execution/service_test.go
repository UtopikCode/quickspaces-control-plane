package execution

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/UtopikCode/quickspaces-control-plane/domain"
	contracts "github.com/UtopikCode/quickspaces-execution-contracts"
)

type testAdapter struct {
	started bool
}

func (a *testAdapter) StartWorkspace(ctx context.Context, workspace contracts.Workspace) (contracts.WorkspaceState, error) {
	a.started = true
	return contracts.WorkspaceStateRunning, nil
}

func (a *testAdapter) StopWorkspace(ctx context.Context, id string) error {
	return nil
}

func (a *testAdapter) GetWorkspaceStatus(ctx context.Context, id string) (contracts.WorkspaceState, error) {
	return contracts.WorkspaceStateRunning, nil
}

func TestExecutionServiceResolvesProviderFromExecutionProfile(t *testing.T) {
	awsAdapter := &testAdapter{}
	registry := NewAdapterRegistry()
	registry.Register("aws", awsAdapter)
	service := NewExecutionService(registry)

	workspace := &domain.Workspace{
		ID:               "ws-1",
		Repo:             "repo",
		Owner:            "owner",
		Ref:              "main",
		DesiredState:     "running",
		ExecutionProfile: json.RawMessage(`{"provider":"aws"}`),
	}

	if err := service.StartWorkspace(context.Background(), workspace); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !awsAdapter.started {
		t.Fatal("expected aws adapter to be used")
	}
}

func TestExecutionServiceFailsWhenProviderUnsupported(t *testing.T) {
	service := NewExecutionService(NewAdapterRegistry())
	workspace := &domain.Workspace{ExecutionProfile: json.RawMessage(`{"provider":"truenas"}`)}

	err := service.StartWorkspace(context.Background(), workspace)
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
}

func TestProviderFromExecutionProfileValidation(t *testing.T) {
	_, err := providerFromExecutionProfile(nil)
	if !errors.Is(err, errMissingProviderInExecutionProfile) {
		t.Fatalf("expected missing provider error, got %v", err)
	}

	_, err = providerFromExecutionProfile(json.RawMessage(`not-json`))
	if err == nil {
		t.Fatal("expected invalid json error")
	}
}
