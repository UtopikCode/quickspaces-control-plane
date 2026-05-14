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
	adapter := &testAdapter{}
	registry := NewAdapterRegistry()
	registry.Register("test", func(_ json.RawMessage) (contracts.ExecutionAdapter, error) {
		return adapter, nil
	})
	service := NewExecutionService(registry, nil)

	workspace := &domain.Workspace{
		ID:               "ws-1",
		Repo:             "repo",
		Owner:            "owner",
		Ref:              "main",
		DesiredState:     "running",
		ExecutionProfile: json.RawMessage(`{"provider":"test"}`),
	}

	if err := service.StartWorkspace(context.Background(), workspace); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !adapter.started {
		t.Fatal("expected resolved adapter to be used")
	}
}

func TestExecutionServiceResolvesProviderCaseInsensitively(t *testing.T) {
	adapter := &testAdapter{}
	registry := NewAdapterRegistry()
	registry.Register("test", func(_ json.RawMessage) (contracts.ExecutionAdapter, error) {
		return adapter, nil
	})
	service := NewExecutionService(registry, nil)

	workspace := &domain.Workspace{
		ExecutionProfile: json.RawMessage(`{"provider":"TeSt"}`),
	}

	if err := service.StartWorkspace(context.Background(), workspace); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !adapter.started {
		t.Fatal("expected resolved adapter to be used for mixed-case provider")
	}
}

func TestExecutionServiceFailsWhenProviderUnsupported(t *testing.T) {
	service := NewExecutionService(NewAdapterRegistry(), nil)
	workspace := &domain.Workspace{ExecutionProfile: json.RawMessage(`{"provider":"truenas"}`)}

	err := service.StartWorkspace(context.Background(), workspace)
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
}

func TestExecutionServiceFailsWhenExecutionProfileMissing(t *testing.T) {
	service := NewExecutionService(NewAdapterRegistry(), nil)
	workspace := &domain.Workspace{ExecutionProfile: nil}

	err := service.StartWorkspace(context.Background(), workspace)
	if err == nil {
		t.Fatal("expected error when execution profile is missing")
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
