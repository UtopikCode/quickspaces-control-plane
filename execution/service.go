package execution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/UtopikCode/quickspaces-control-plane/domain"
	contracts "github.com/UtopikCode/quickspaces-execution-contracts"
)

type ExecutionService struct {
	registry AdapterResolver
}

type AdapterResolver interface {
	Resolve(provider string) (contracts.ExecutionAdapter, error)
}

func NewExecutionService(registry AdapterResolver) *ExecutionService {
	return &ExecutionService{registry: registry}
}

func ValidateExecutionProfile(raw json.RawMessage) error {
	_, err := providerFromExecutionProfile(raw)
	return err
}

func (s *ExecutionService) StartWorkspace(ctx context.Context, workspace *domain.Workspace) error {
	adapter, err := s.adapterForWorkspace(workspace)
	if err != nil {
		return err
	}
	ws, err := toContractsWorkspace(workspace)
	if err != nil {
		return err
	}
	_, err = adapter.StartWorkspace(ctx, ws)
	return err
}

func (s *ExecutionService) StopWorkspace(ctx context.Context, workspace *domain.Workspace) error {
	adapter, err := s.adapterForWorkspace(workspace)
	if err != nil {
		return err
	}
	return adapter.StopWorkspace(ctx, workspace.ID)
}

func (s *ExecutionService) GetWorkspaceStatus(ctx context.Context, workspace *domain.Workspace) (string, error) {
	adapter, err := s.adapterForWorkspace(workspace)
	if err != nil {
		return "", err
	}
	status, err := adapter.GetWorkspaceStatus(ctx, workspace.ID)
	return string(status), err
}

func toContractsWorkspace(workspace *domain.Workspace) (contracts.Workspace, error) {
	executionProfile, err := parseExecutionProfile(workspace.ExecutionProfile)
	if err != nil {
		return contracts.Workspace{}, err
	}
	return contracts.Workspace{
		ID:               workspace.ID,
		Repo:             workspace.Repo,
		Owner:            workspace.Owner,
		Ref:              workspace.Ref,
		ExecutionProfile: executionProfile,
	}, nil
}

func (s *ExecutionService) adapterForWorkspace(workspace *domain.Workspace) (contracts.ExecutionAdapter, error) {
	provider, err := providerFromExecutionProfile(workspace.ExecutionProfile)
	if err != nil {
		return nil, err
	}

	adapter, err := s.registry.Resolve(provider)
	if err != nil {
		return nil, fmt.Errorf("resolve execution adapter %q: %w", provider, err)
	}
	return adapter, nil
}

var errMissingProviderInExecutionProfile = errors.New("executionProfile.provider is required")

func parseExecutionProfile(raw json.RawMessage) (contracts.ExecutionProfile, error) {
	if len(raw) == 0 {
		return contracts.ExecutionProfile{}, errMissingProviderInExecutionProfile
	}

	var profile contracts.ExecutionProfile
	if err := json.Unmarshal(raw, &profile); err != nil {
		return contracts.ExecutionProfile{}, fmt.Errorf("invalid executionProfile: %w", err)
	}

	if strings.TrimSpace(profile.Provider) == "" {
		return contracts.ExecutionProfile{}, errMissingProviderInExecutionProfile
	}

	return profile, nil
}

func providerFromExecutionProfile(raw json.RawMessage) (string, error) {
	profile, err := parseExecutionProfile(raw)
	if err != nil {
		return "", err
	}
	return strings.ToLower(strings.TrimSpace(profile.Provider)), nil
}
