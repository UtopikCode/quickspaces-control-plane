package auth

import (
	"context"
	"errors"
	"strings"

	githubclient "github.com/UtopikCode/quickspaces-control-plane/internal/infrastructure/github"
)

type AccessRule struct {
	SubjectType string `json:"subjectType"`
	SubjectID   string `json:"subjectId"`
	Role        string `json:"role"`
}

type AccessRuleRepository interface {
	List(ctx context.Context) ([]*AccessRule, error)
	Upsert(ctx context.Context, subjectType, subjectID, role string) error
	Delete(ctx context.Context, subjectType, subjectID string) error
}

type Service struct {
	repo          AccessRuleRepository
	initialAdmins map[string]struct{}
}

func NewService(repo AccessRuleRepository, initialAdmins []string) *Service {
	adminSet := make(map[string]struct{}, len(initialAdmins))
	for _, admin := range initialAdmins {
		admin = strings.ToLower(strings.TrimSpace(admin))
		if admin != "" {
			adminSet[admin] = struct{}{}
		}
	}
	return &Service{repo: repo, initialAdmins: adminSet}
}

func (s *Service) Authorize(ctx context.Context, user githubclient.GithubUser, orgs []string, teams []githubclient.GithubTeam) (bool, string, error) {
	if user.Login == "" {
		return false, "", errors.New("invalid user identity")
	}

	normalizedUser := strings.ToLower(strings.TrimSpace(user.Login))
	orgSet := make(map[string]struct{}, len(orgs))
	for _, org := range orgs {
		orgSet[strings.ToLower(strings.TrimSpace(org))] = struct{}{}
	}
	teamSet := make(map[string]struct{}, len(teams))
	for _, team := range teams {
		teamSet[strings.ToLower(strings.TrimSpace(team.Org)+"/"+strings.TrimSpace(team.Name))] = struct{}{}
	}

	rules, err := s.repo.List(ctx)
	if err != nil {
		return false, "", err
	}
	if len(rules) == 0 {
		if len(s.initialAdmins) == 0 {
			return false, "", nil
		}
		if _, ok := s.initialAdmins[normalizedUser]; ok {
			return true, "admin", nil
		}
		return false, "", nil
	}

	bestRole := ""
	for _, rule := range rules {
		ruleType := strings.ToLower(strings.TrimSpace(rule.SubjectType))
		ruleID := strings.ToLower(strings.TrimSpace(rule.SubjectID))
		ruleRole := strings.ToLower(strings.TrimSpace(rule.Role))
		if ruleID == "" {
			continue
		}
		if ruleRole != "user" && ruleRole != "admin" {
			continue
		}

		switch ruleType {
		case "user":
			if ruleID == normalizedUser {
				bestRole = selectBestRole(bestRole, ruleRole)
			}
		case "org":
			if _, ok := orgSet[ruleID]; ok {
				bestRole = selectBestRole(bestRole, ruleRole)
			}
		case "team":
			if _, ok := teamSet[ruleID]; ok {
				bestRole = selectBestRole(bestRole, ruleRole)
			}
		}
	}

	if bestRole == "" {
		return false, "", nil
	}
	return true, bestRole, nil
}

func (s *Service) GrantAccess(ctx context.Context, subjectType, subjectID, role string) error {
	subjectType = strings.ToLower(strings.TrimSpace(subjectType))
	subjectID = strings.ToLower(strings.TrimSpace(subjectID))
	role = strings.ToLower(strings.TrimSpace(role))

	switch subjectType {
	case "user", "org", "team":
		// valid
	default:
		return errors.New("subjectType must be one of user, org, team")
	}

	if subjectID == "" {
		return errors.New("subjectId is required")
	}

	switch role {
	case "user", "admin":
		// valid
	default:
		return errors.New("role must be user or admin")
	}

	return s.repo.Upsert(ctx, subjectType, subjectID, role)
}

func (s *Service) ListAccess(ctx context.Context) ([]*AccessRule, error) {
	return s.repo.List(ctx)
}

func (s *Service) RemoveAccess(ctx context.Context, subjectType, subjectID string) error {
	subjectType = strings.ToLower(strings.TrimSpace(subjectType))
	subjectID = strings.ToLower(strings.TrimSpace(subjectID))

	switch subjectType {
	case "user", "org", "team":
		// valid
	default:
		return errors.New("subjectType must be one of user, org, team")
	}

	if subjectID == "" {
		return errors.New("subjectId is required")
	}

	return s.repo.Delete(ctx, subjectType, subjectID)
}

func selectBestRole(current, candidate string) string {
	if current == "" {
		return candidate
	}
	if candidate == "admin" {
		return candidate
	}
	return current
}
