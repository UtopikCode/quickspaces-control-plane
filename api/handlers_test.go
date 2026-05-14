package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/UtopikCode/quickspaces-control-plane/application"
	"github.com/UtopikCode/quickspaces-control-plane/domain"
	"github.com/UtopikCode/quickspaces-control-plane/execution"
	"github.com/UtopikCode/quickspaces-control-plane/internal/application/auth"
	githubclient "github.com/UtopikCode/quickspaces-control-plane/internal/infrastructure/github"
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

func (a *testAdapter) StartWorkspace(ctx context.Context, workspace contracts.Workspace) (contracts.WorkspaceState, error) {
	return contracts.WorkspaceStateRunning, nil
}

func (a *testAdapter) StopWorkspace(ctx context.Context, id string) error {
	return nil
}

type fakeAccessRuleRepo struct {
	rules []*auth.AccessRule
}

func (f *fakeAccessRuleRepo) List(ctx context.Context) ([]*auth.AccessRule, error) {
	return f.rules, nil
}

func (f *fakeAccessRuleRepo) Upsert(ctx context.Context, subjectType, subjectID, role string) error {
	for _, rule := range f.rules {
		if rule.SubjectType == subjectType && rule.SubjectID == subjectID {
			rule.Role = role
			return nil
		}
	}
	f.rules = append(f.rules, &auth.AccessRule{SubjectType: subjectType, SubjectID: subjectID, Role: role})
	return nil
}

func (f *fakeAccessRuleRepo) Delete(ctx context.Context, subjectType, subjectID string) error {
	filtered := make([]*auth.AccessRule, 0, len(f.rules))
	for _, rule := range f.rules {
		if rule.SubjectType == subjectType && rule.SubjectID == subjectID {
			continue
		}
		filtered = append(filtered, rule)
	}
	f.rules = filtered
	return nil
}

func (a *testAdapter) GetWorkspaceStatus(ctx context.Context, id string) (contracts.WorkspaceState, error) {
	return contracts.WorkspaceStateRunning, nil
}

type mockGitHubClient struct{}

func (m *mockGitHubClient) AuthorizeURL() string {
	return "https://github.com/login/oauth/authorize?client_id=test"
}

func (m *mockGitHubClient) ExchangeCode(code string, codeVerifier string) (string, error) {
	return "test-token-" + code, nil
}

func (m *mockGitHubClient) GetUser(token string) (githubclient.GithubUser, error) {
	return githubclient.GithubUser{Login: "testuser", ID: 123}, nil
}

func (m *mockGitHubClient) GetUserOrgs(token string) ([]string, error) {
	return []string{"test-org"}, nil
}

func (m *mockGitHubClient) GetUserTeams(token string) ([]githubclient.GithubTeam, error) {
	return []githubclient.GithubTeam{{Org: "test-org", Name: "test-team"}}, nil
}

func TestHealthEndpoint(t *testing.T) {
	repo := newTestRepo()
	registry := execution.NewAdapterRegistry()
	registry.Register("truenas", &testAdapter{})
	service := application.NewWorkspaceService(repo, execution.NewExecutionService(registry))
	authService := auth.NewService(&fakeAccessRuleRepo{rules: []*auth.AccessRule{{SubjectType: "user", SubjectID: "testuser", Role: "admin"}}}, nil)
	h := NewHandler(service, authService, &mockGitHubClient{})
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
	registry := execution.NewAdapterRegistry()
	registry.Register("truenas", &testAdapter{})
	service := application.NewWorkspaceService(repo, execution.NewExecutionService(registry))
	authService := auth.NewService(&fakeAccessRuleRepo{rules: []*auth.AccessRule{{SubjectType: "user", SubjectID: "testuser", Role: "admin"}}}, nil)
	h := NewHandler(service, authService, &mockGitHubClient{})
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
	createReq.Header.Set("Authorization", "Bearer testtoken")
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
	startReq.Header.Set("Authorization", "Bearer testtoken")
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
