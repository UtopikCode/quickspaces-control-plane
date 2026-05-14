package postgres

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/UtopikCode/quickspaces-control-plane/domain"
	entclient "github.com/UtopikCode/quickspaces-control-plane/ent"
	"github.com/UtopikCode/quickspaces-control-plane/ent/workspace"
)

type WorkspaceRepository struct {
	client *entclient.Client
}

func NewWorkspaceRepository(client *entclient.Client) *WorkspaceRepository {
	return &WorkspaceRepository{client: client}
}

func toDomainWorkspace(entity *entclient.Workspace) *domain.Workspace {
	if entity == nil {
		return nil
	}
	return &domain.Workspace{
		ID:               entity.ID,
		Repo:             entity.Repo,
		Owner:            entity.Owner,
		Ref:              entity.Ref,
		DesiredState:     entity.DesiredState,
		ActualState:      entity.ActualState,
		ExecutionProfile: entity.ExecutionProfile,
		CreatedAt:        entity.CreatedAt,
		UpdatedAt:        entity.UpdatedAt,
	}
}

func (r *WorkspaceRepository) Create(ctx context.Context, workspaceModel *domain.Workspace) error {
	_, err := r.client.Workspace.Create().
		SetID(workspaceModel.ID).
		SetRepo(workspaceModel.Repo).
		SetOwner(workspaceModel.Owner).
		SetRef(workspaceModel.Ref).
		SetDesiredState(workspaceModel.DesiredState).
		SetActualState(workspaceModel.ActualState).
		SetExecutionProfile(workspaceModel.ExecutionProfile).
		SetCreatedAt(workspaceModel.CreatedAt).
		SetUpdatedAt(workspaceModel.UpdatedAt).
		Save(ctx)
	return err
}

func (r *WorkspaceRepository) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	entity, err := r.client.Workspace.Get(ctx, id)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domain.ErrWorkspaceNotFound
		}
		return nil, err
	}
	return toDomainWorkspace(entity), nil
}

func (r *WorkspaceRepository) List(ctx context.Context) ([]*domain.Workspace, error) {
	entities, err := r.client.Workspace.Query().Order(workspace.ByCreatedAt(sql.OrderDesc())).All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Workspace, 0, len(entities))
	for _, entity := range entities {
		result = append(result, toDomainWorkspace(entity))
	}
	return result, nil
}

func (r *WorkspaceRepository) UpdateDesiredState(ctx context.Context, id, desiredState string, updatedAt time.Time) error {
	_, err := r.client.Workspace.UpdateOneID(id).
		SetDesiredState(desiredState).
		SetUpdatedAt(updatedAt).
		Save(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return domain.ErrWorkspaceNotFound
		}
	}
	return err
}

func (r *WorkspaceRepository) UpdateActualState(ctx context.Context, id, actualState string, updatedAt time.Time) error {
	_, err := r.client.Workspace.UpdateOneID(id).
		SetActualState(actualState).
		SetUpdatedAt(updatedAt).
		Save(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return domain.ErrWorkspaceNotFound
		}
	}
	return err
}
