package auth

import (
	"context"
	"testing"

	githubclient "github.com/UtopikCode/quickspaces-control-plane/internal/infrastructure/github"
)

type fakeAccessRuleRepo struct {
	rules []*AccessRule
}

func (f *fakeAccessRuleRepo) List(ctx context.Context) ([]*AccessRule, error) {
	return f.rules, nil
}

func (f *fakeAccessRuleRepo) Upsert(ctx context.Context, subjectType, subjectID, role string) error {
	for _, rule := range f.rules {
		if rule.SubjectType == subjectType && rule.SubjectID == subjectID {
			rule.Role = role
			return nil
		}
	}
	f.rules = append(f.rules, &AccessRule{SubjectType: subjectType, SubjectID: subjectID, Role: role})
	return nil
}

func (f *fakeAccessRuleRepo) Delete(ctx context.Context, subjectType, subjectID string) error {
	filtered := make([]*AccessRule, 0, len(f.rules))
	for _, rule := range f.rules {
		if rule.SubjectType == subjectType && rule.SubjectID == subjectID {
			continue
		}
		filtered = append(filtered, rule)
	}
	f.rules = filtered
	return nil
}

func TestAuthorizeMatchesUser(t *testing.T) {
	repo := &fakeAccessRuleRepo{rules: []*AccessRule{{SubjectType: "user", SubjectID: "alice", Role: "admin"}}}
	service := NewService(repo, nil)

	allowed, role, err := service.Authorize(context.Background(), githubclient.GithubUser{Login: "alice"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed || role != "admin" {
		t.Fatalf("expected admin user match, got allowed=%v role=%q", allowed, role)
	}
}

func TestAuthorizeMatchesOrg(t *testing.T) {
	repo := &fakeAccessRuleRepo{rules: []*AccessRule{{SubjectType: "org", SubjectID: "acme", Role: "user"}}}
	service := NewService(repo, nil)

	allowed, role, err := service.Authorize(context.Background(), githubclient.GithubUser{Login: "bob"}, []string{"acme"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed || role != "user" {
		t.Fatalf("expected org match, got allowed=%v role=%q", allowed, role)
	}
}

func TestAuthorizeMatchesTeam(t *testing.T) {
	repo := &fakeAccessRuleRepo{rules: []*AccessRule{{SubjectType: "team", SubjectID: "acme/developers", Role: "admin"}}}
	service := NewService(repo, nil)

	allowed, role, err := service.Authorize(context.Background(), githubclient.GithubUser{Login: "carol"}, nil, []githubclient.GithubTeam{{Org: "acme", Name: "developers"}})
	if err != nil {
		t.Fatal(err)
	}
	if !allowed || role != "admin" {
		t.Fatalf("expected team admin match, got allowed=%v role=%q", allowed, role)
	}
}

func TestAuthorizeRejectsUnknown(t *testing.T) {
	repo := &fakeAccessRuleRepo{rules: []*AccessRule{{SubjectType: "user", SubjectID: "alice", Role: "admin"}}}
	service := NewService(repo, nil)

	allowed, _, err := service.Authorize(context.Background(), githubclient.GithubUser{Login: "bob"}, []string{"acme"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected access denied for unknown user")
	}
}

func TestAuthorizeAllowsInitialAdminWhenNoRules(t *testing.T) {
	repo := &fakeAccessRuleRepo{}
	service := NewService(repo, []string{"alice"})

	allowed, role, err := service.Authorize(context.Background(), githubclient.GithubUser{Login: "alice"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed || role != "admin" {
		t.Fatalf("expected initial admin to be granted, got allowed=%v role=%q", allowed, role)
	}
}

func TestAuthorizeRejectsWhenNoRulesAndInitialAdminNotConfigured(t *testing.T) {
	repo := &fakeAccessRuleRepo{}
	service := NewService(repo, nil)

	allowed, _, err := service.Authorize(context.Background(), githubclient.GithubUser{Login: "alice"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected access denied when no rules exist and no initial admin configured")
	}
}

func TestAuthorizeBootstrapsInitialRulesWhenNoExistingRules(t *testing.T) {
	repo := &fakeAccessRuleRepo{}
	service := NewService(repo, []string{"alice", "org:acme", "team:acme/developers"})

	allowed, role, err := service.Authorize(context.Background(), githubclient.GithubUser{Login: "alice"}, []string{"acme"}, []githubclient.GithubTeam{{Org: "acme", Name: "developers"}})
	if err != nil {
		t.Fatal(err)
	}
	if !allowed || role != "admin" {
		t.Fatalf("expected admin access after bootstrap, got allowed=%v role=%q", allowed, role)
	}
	if len(repo.rules) != 3 {
		t.Fatalf("expected 3 bootstrap rules to be created, got %d", len(repo.rules))
	}
}
