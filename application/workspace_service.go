package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/UtopikCode/quickspaces-control-plane/domain"
	"github.com/UtopikCode/quickspaces-control-plane/execution"
)

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *domain.Workspace) error
	GetByID(ctx context.Context, id string) (*domain.Workspace, error)
	List(ctx context.Context) ([]*domain.Workspace, error)
	UpdateDesiredState(ctx context.Context, id, desiredState string, updatedAt time.Time) error
	UpdateActualState(ctx context.Context, id, actualState string, updatedAt time.Time) error
}

type WorkspaceService struct {
	repo    WorkspaceRepository
	execSvc *execution.ExecutionService
}

func NewWorkspaceService(repo WorkspaceRepository, execSvc *execution.ExecutionService) *WorkspaceService {
	return &WorkspaceService{repo: repo, execSvc: execSvc}
}

type CreateWorkspaceRequest struct {
	Repo             string                  `json:"repo"`
	Owner            string                  `json:"owner"`
	Ref              string                  `json:"ref"`
	ExecutionProfile domain.ExecutionProfile `json:"executionProfile"`
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, request CreateWorkspaceRequest) (*domain.Workspace, error) {
	if request.Repo == "" || request.Owner == "" || request.Ref == "" {
		return nil, errors.New("repo, owner, and ref are required")
	}

	now := time.Now().UTC()
	workspace := &domain.Workspace{
		ID:               generateID(),
		Repo:             request.Repo,
		Owner:            request.Owner,
		Ref:              request.Ref,
		DesiredState:     "stopped",
		ActualState:      "stopped",
		ExecutionProfile: request.ExecutionProfile,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (s *WorkspaceService) ListWorkspaces(ctx context.Context) ([]*domain.Workspace, error) {
	return s.repo.List(ctx)
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context, id string) (*domain.Workspace, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *WorkspaceService) StartWorkspace(ctx context.Context, id string) (*domain.Workspace, error) {
	workspace, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	workspace.DesiredState = "running"
	workspace.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateDesiredState(ctx, id, workspace.DesiredState, workspace.UpdatedAt); err != nil {
		return nil, err
	}

	if err := s.execSvc.StartWorkspace(ctx, workspace); err != nil {
		return workspace, err
	}

	return workspace, nil
}

func (s *WorkspaceService) StopWorkspace(ctx context.Context, id string) (*domain.Workspace, error) {
	workspace, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	workspace.DesiredState = "stopped"
	workspace.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateDesiredState(ctx, id, workspace.DesiredState, workspace.UpdatedAt); err != nil {
		return nil, err
	}

	if err := s.execSvc.StopWorkspace(ctx, workspace); err != nil {
		return workspace, err
	}

	return workspace, nil
}

func (s *WorkspaceService) ReconcileWorkspace(ctx context.Context, id string) (*domain.Workspace, error) {
	workspace, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	status, err := s.execSvc.GetWorkspaceStatus(ctx, workspace)
	if err != nil {
		return nil, err
	}

	workspace.ActualState = status
	workspace.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateActualState(ctx, id, workspace.ActualState, workspace.UpdatedAt); err != nil {
		return nil, err
	}

	return workspace, nil
}

func generateID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
