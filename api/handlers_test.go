package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/UtopikCode/quickspaces-control-plane/application"
	"github.com/UtopikCode/quickspaces-control-plane/domain"
	"github.com/UtopikCode/quickspaces-control-plane/execution"
	contracts "github.com/UtopikCode/quickspaces-execution-contracts"
)

type testRepo struct {
	store map[string]*domain.Workspace
}

func newTestRepo() *testRepo {
	return &testRepo{store: make(map[string]*domain.Workspace)}
}

func (r *testRepo) Create(ctx context.Context, workspace *domain.Workspace) error {
	r.store[workspace.ID] = workspace
	return nil
}

func (r *testRepo) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	workspace, ok := r.store[id]
	if !ok {
		return nil, domain.ErrWorkspaceNotFound
	}
	return workspace, nil
}

func (r *testRepo) List(ctx context.Context) ([]*domain.Workspace, error) {
	result := make([]*domain.Workspace, 0, len(r.store))
	for _, workspace := range r.store {
		result = append(result, workspace)
	}
	return result, nil
}

func (r *testRepo) UpdateDesiredState(ctx context.Context, id, desiredState string, updatedAt time.Time) error {
	workspace, ok := r.store[id]
	if !ok {
		return domain.ErrWorkspaceNotFound
	}
	workspace.DesiredState = desiredState
	workspace.UpdatedAt = updatedAt
	return nil
}

func (r *testRepo) UpdateActualState(ctx context.Context, id, actualState string, updatedAt time.Time) error {
	workspace, ok := r.store[id]
	if !ok {
		return domain.ErrWorkspaceNotFound
	}
	workspace.ActualState = actualState
	workspace.UpdatedAt = updatedAt
	return nil
}

type testAdapter struct{}

func (a *testAdapter) StartWorkspace(ctx context.Context, workspace contracts.WorkspaceSpec) error {
	return nil
}

func (a *testAdapter) StopWorkspace(ctx context.Context, workspace contracts.WorkspaceSpec) error {
	return nil
}

func (a *testAdapter) GetWorkspaceStatus(ctx context.Context, workspace contracts.WorkspaceSpec) (string, error) {
	if workspace.DesiredState == "running" {
		return "running", nil
	}
	return "stopped", nil
}

func TestHealthEndpoint(t *testing.T) {
	repo := newTestRepo()
	service := application.NewWorkspaceService(repo, execution.NewExecutionService(&testAdapter{}))
	h := NewHandler(service)
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestCreateAndStartWorkspace(t *testing.T) {
	repo := newTestRepo()
	service := application.NewWorkspaceService(repo, execution.NewExecutionService(&testAdapter{}))
	h := NewHandler(service)
	router := NewRouter(h)

	payload := map[string]interface{}{
		"repo":             "github.com/example/repo",
		"owner":            "team-a",
		"ref":              "main",
		"executionProfile": map[string]string{"provider": "truenas"},
	}
	body, _ := json.Marshal(payload)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createRes := httptest.NewRecorder()
	router.ServeHTTP(createRes, createReq)

	if createRes.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createRes.Code)
	}

	var created domain.Workspace
	if err := json.NewDecoder(createRes.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+created.ID+"/start", nil)
	startRes := httptest.NewRecorder()
	router.ServeHTTP(startRes, startReq)

	if startRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", startRes.Code)
	}

	var started domain.Workspace
	if err := json.NewDecoder(startRes.Body).Decode(&started); err != nil {
		t.Fatal(err)
	}
	if started.DesiredState != "running" {
		t.Fatalf("expected running desired state, got %s", started.DesiredState)
	}
}
func TestSwaggerUIEndpoint(t *testing.T) {
	repo := newTestRepo()
	service := application.NewWorkspaceService(repo, execution.NewExecutionService(&testAdapter{}))
	h := NewHandler(service)
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	if ct := res.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("expected HTML response, got %s", ct)
	}
}

func TestSwaggerDocJSONEndpoint(t *testing.T) {
	repo := newTestRepo()
	service := application.NewWorkspaceService(repo, execution.NewExecutionService(&testAdapter{}))
	h := NewHandler(service)
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	if ct := res.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("expected JSON response, got %s", ct)
	}

	var spec map[string]any
	if err := json.NewDecoder(res.Body).Decode(&spec); err != nil {
		t.Fatalf("failed to decode Swagger JSON: %v", err)
	}
	if specVersion, ok := spec["swagger"].(string); !ok || specVersion != "2.0" {
		t.Fatalf("expected swagger version 2.0, got %v", spec["swagger"])
	}
}
