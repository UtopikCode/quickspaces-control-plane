package application

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/UtopikCode/quickspaces-control-plane/domain"
	"github.com/UtopikCode/quickspaces-control-plane/execution"
	contracts "github.com/UtopikCode/quickspaces-execution-contracts"
)

type mockRepo struct {
	store map[string]*domain.Workspace
}

func newMockRepo() *mockRepo {
	return &mockRepo{store: make(map[string]*domain.Workspace)}
}

func (r *mockRepo) Create(ctx context.Context, workspace *domain.Workspace) error {
	r.store[workspace.ID] = workspace
	return nil
}

func (r *mockRepo) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	workspace, ok := r.store[id]
	if !ok {
		return nil, domain.ErrWorkspaceNotFound
	}
	return workspace, nil
}

func (r *mockRepo) List(ctx context.Context) ([]*domain.Workspace, error) {
	result := make([]*domain.Workspace, 0, len(r.store))
	for _, workspace := range r.store {
		result = append(result, workspace)
	}
	return result, nil
}

func (r *mockRepo) UpdateDesiredState(ctx context.Context, id, desiredState string, updatedAt time.Time) error {
	workspace, ok := r.store[id]
	if !ok {
		return domain.ErrWorkspaceNotFound
	}
	workspace.DesiredState = desiredState
	workspace.UpdatedAt = updatedAt
	return nil
}

func (r *mockRepo) UpdateActualState(ctx context.Context, id, actualState string, updatedAt time.Time) error {
	workspace, ok := r.store[id]
	if !ok {
		return domain.ErrWorkspaceNotFound
	}
	workspace.ActualState = actualState
	workspace.UpdatedAt = updatedAt
	return nil
}

type mockHostRepo struct {
	hosts map[string]*domain.ExecutionHost
}

func newMockHostRepo() *mockHostRepo {
	return &mockHostRepo{hosts: make(map[string]*domain.ExecutionHost)}
}

func (r *mockHostRepo) GetByID(ctx context.Context, id string) (*domain.ExecutionHost, error) {
	host, ok := r.hosts[id]
	if !ok {
		return nil, domain.ErrHostNotFound
	}
	return host, nil
}

func (r *mockHostRepo) Create(ctx context.Context, host *domain.ExecutionHost) error {
	r.hosts[host.ID] = host
	return nil
}

func (r *mockHostRepo) List(ctx context.Context) ([]*domain.ExecutionHost, error) {
	result := make([]*domain.ExecutionHost, 0, len(r.hosts))
	for _, host := range r.hosts {
		result = append(result, host)
	}
	return result, nil
}

type mockAdapter struct {
	started bool
	stopped bool
}

func (m *mockAdapter) StartWorkspace(ctx context.Context, workspace contracts.Workspace) (contracts.WorkspaceState, error) {
	m.started = true
	return contracts.WorkspaceStateRunning, nil
}

func (m *mockAdapter) StopWorkspace(ctx context.Context, id string) error {
	m.stopped = true
	return nil
}

func (m *mockAdapter) GetWorkspaceStatus(ctx context.Context, id string) (contracts.WorkspaceState, error) {
	return contracts.WorkspaceStateStopped, nil
}

func TestCreateWorkspace(t *testing.T) {
	repo := newMockRepo()
	hostRepo := newMockHostRepo()
	hostRepo.hosts["host-1"] = &domain.ExecutionHost{ID: "host-1", Name: "trueNAS host", Adapter: "truenas", Config: json.RawMessage(`{}`), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	adapter := &mockAdapter{}
	registry := execution.NewAdapterRegistry()
	registry.Register("truenas", func(_ json.RawMessage) (contracts.ExecutionAdapter, error) {
		return adapter, nil
	})
	service := NewWorkspaceService(repo, hostRepo, execution.NewExecutionService(registry, hostRepo))

	workspace, err := service.CreateWorkspace(context.Background(), CreateWorkspaceRequest{
		Repo:             "github.com/example/repo",
		Owner:            "owner",
		Ref:              "main",
		HostID:           "host-1",
		ExecutionProfile: domain.ExecutionProfile([]byte(`{"runtimeConfig":{"image":"ubuntu:24.04"}}`)),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if workspace.DesiredState != "stopped" {
		t.Fatalf("expected desired state stopped, got %s", workspace.DesiredState)
	}
}

func TestStartStopAndReconcile(t *testing.T) {
	repo := newMockRepo()
	hostRepo := newMockHostRepo()
	hostRepo.hosts["host-1"] = &domain.ExecutionHost{ID: "host-1", Name: "trueNAS host", Adapter: "truenas", Config: json.RawMessage(`{}`), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	adapter := &mockAdapter{}
	registry := execution.NewAdapterRegistry()
	registry.Register("truenas", func(_ json.RawMessage) (contracts.ExecutionAdapter, error) {
		return adapter, nil
	})
	service := NewWorkspaceService(repo, hostRepo, execution.NewExecutionService(registry, hostRepo))

	workspace, err := service.CreateWorkspace(context.Background(), CreateWorkspaceRequest{
		Repo:             "github.com/example/repo",
		Owner:            "owner",
		Ref:              "main",
		HostID:           "host-1",
		ExecutionProfile: domain.ExecutionProfile([]byte(`{"runtimeConfig":{"image":"ubuntu:24.04"}}`)),
	})
	if err != nil {
		t.Fatal(err)
	}

	started, err := service.StartWorkspace(context.Background(), workspace.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !adapter.started {
		t.Fatal("expected adapter StartWorkspace called")
	}
	if started.DesiredState != "running" {
		t.Fatalf("expected running desired state, got %s", started.DesiredState)
	}

	stopped, err := service.StopWorkspace(context.Background(), workspace.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !adapter.stopped {
		t.Fatal("expected adapter StopWorkspace called")
	}
	if stopped.DesiredState != "stopped" {
		t.Fatalf("expected stopped desired state, got %s", stopped.DesiredState)
	}

	reconciled, err := service.ReconcileWorkspace(context.Background(), workspace.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reconciled.ActualState != "stopped" {
		t.Fatalf("expected actual state stopped, got %s", reconciled.ActualState)
	}
}

func TestGetWorkspaceNotFound(t *testing.T) {
	repo := newMockRepo()
	hostRepo := newMockHostRepo()
	adapter := &mockAdapter{}
	registry := execution.NewAdapterRegistry()
	registry.Register("truenas", func(_ json.RawMessage) (contracts.ExecutionAdapter, error) {
		return adapter, nil
	})
	service := NewWorkspaceService(repo, hostRepo, execution.NewExecutionService(registry, hostRepo))

	_, err := service.GetWorkspace(context.Background(), "missing")
	if !errors.Is(err, domain.ErrWorkspaceNotFound) {
		t.Fatalf("expected ErrWorkspaceNotFound, got %v", err)
	}
}

func TestListWorkspaces(t *testing.T) {
	repo := newMockRepo()
	hostRepo := newMockHostRepo()
	hostRepo.hosts["host-1"] = &domain.ExecutionHost{ID: "host-1", Name: "trueNAS host", Adapter: "truenas", Config: json.RawMessage(`{}`), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	adapter := &mockAdapter{}
	registry := execution.NewAdapterRegistry()
	registry.Register("truenas", func(_ json.RawMessage) (contracts.ExecutionAdapter, error) {
		return adapter, nil
	})
	service := NewWorkspaceService(repo, hostRepo, execution.NewExecutionService(registry, hostRepo))

	_, err := service.CreateWorkspace(context.Background(), CreateWorkspaceRequest{
		Repo:             "github.com/example/repo",
		Owner:            "owner",
		Ref:              "main",
		HostID:           "host-1",
		ExecutionProfile: domain.ExecutionProfile([]byte(`{"runtimeConfig":{"image":"ubuntu:24.04"}}`)),
	})
	if err != nil {
		t.Fatal(err)
	}

	list, err := service.ListWorkspaces(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(list))
	}
	if !reflect.DeepEqual(list[0].Repo, "github.com/example/repo") {
		t.Fatalf("unexpected workspace repo: %s", list[0].Repo)
	}
}
