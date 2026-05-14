package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	githubclient "github.com/UtopikCode/quickspaces-control-plane/internal/infrastructure/github"
)

type fakeGitHubClient struct {
	user  githubclient.GithubUser
	orgs  []string
	teams []githubclient.GithubTeam
	err   error
}

func (f *fakeGitHubClient) AuthorizeURL() string {
	return "https://github.com/login/oauth/authorize"
}

func (f *fakeGitHubClient) ExchangeCode(code string, codeVerifier string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return "fake-token", nil
}

func (f *fakeGitHubClient) GetUser(token string) (githubclient.GithubUser, error) {
	if f.err != nil {
		return githubclient.GithubUser{}, f.err
	}
	return f.user, nil
}

func (f *fakeGitHubClient) GetUserOrgs(token string) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.orgs, nil
}

func (f *fakeGitHubClient) GetUserTeams(token string) ([]githubclient.GithubTeam, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.teams, nil
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	client := &fakeGitHubClient{err: errors.New("unauthorized")}
	middleware := NewAuthMiddleware(client)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	res := httptest.NewRecorder()

	middleware.Handler(next).ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}
}

func TestAuthMiddlewareValidToken(t *testing.T) {
	client := &fakeGitHubClient{user: githubclient.GithubUser{Login: "bob"}}
	middleware := NewAuthMiddleware(client)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)
	req.Header.Set("Authorization", "Bearer valid")
	res := httptest.NewRecorder()

	middleware.Handler(next).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestAuthMiddlewareSuccess(t *testing.T) {
	client := &fakeGitHubClient{user: githubclient.GithubUser{Login: "alice"}}
	middleware := NewAuthMiddleware(client)

	var gotUser githubclient.GithubUser
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			t.Fatal("expected authenticated user in context")
		}
		gotUser = user
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)
	req.Header.Set("Authorization", "Bearer valid")
	res := httptest.NewRecorder()

	middleware.Handler(next).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	if gotUser.Login != "alice" {
		t.Fatalf("unexpected context user: %v", gotUser)
	}
}

func TestRequireRoleMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)
	req = req.WithContext(WithRole(context.Background(), "user"))
	res := httptest.NewRecorder()

	handler := RequireRole("admin")(next)
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res.Code)
	}
}
