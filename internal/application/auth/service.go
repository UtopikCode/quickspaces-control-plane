package auth

import (
	"context"
	"errors"
	"log"
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
	repo         AccessRuleRepository
	initialRules []*AccessRule
}

func NewService(repo AccessRuleRepository, initialAccessRules []string) *Service {
	rules := make([]*AccessRule, 0, len(initialAccessRules))
	for _, spec := range initialAccessRules {
		if rule, ok := parseAccessRuleSpec(spec); ok {
			rules = append(rules, rule)
		}
	}
	return &Service{repo: repo, initialRules: rules}
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
	if len(rules) == 0 && len(s.initialRules) > 0 {
		if err := s.bootstrapInitialRules(ctx); err != nil {
			return false, "", err
		}
		rules = append([]*AccessRule(nil), s.initialRules...)
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

func (s *Service) bootstrapInitialRules(ctx context.Context) error {
	for _, rule := range s.initialRules {
		if err := s.repo.Upsert(ctx, rule.SubjectType, rule.SubjectID, rule.Role); err != nil {
			return err
		}
	}
	log.Printf("[auth] bootstrapped %d initial access rule(s) from ADMIN_USERS", len(s.initialRules))
	return nil
}

func parseAccessRuleSpec(spec string) (*AccessRule, bool) {
	rule := strings.ToLower(strings.TrimSpace(spec))
	if rule == "" {
		return nil, false
	}

	subjectType := "user"
	subjectID := rule
	if strings.Contains(rule, ":") {
		parts := strings.SplitN(rule, ":", 2)
		subjectType = strings.TrimSpace(parts[0])
		subjectID = strings.TrimSpace(parts[1])
	}

	if subjectID == "" {
		return nil, false
	}

	switch subjectType {
	case "user", "org", "team":
		return &AccessRule{SubjectType: subjectType, SubjectID: subjectID, Role: "admin"}, true
	default:
		return nil, false
	}
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
