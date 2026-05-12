package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/UtopikCode/quickspaces-control-plane/domain"
)

type WorkspaceRepository struct {
	db *sql.DB
}

func NewWorkspaceRepository(db *sql.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

func (r *WorkspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) error {
	const query = `INSERT INTO workspaces (
		id,
		repo,
		owner,
		ref,
		desired_state,
		actual_state,
		execution_profile,
		created_at,
		updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`

	_, err := r.db.ExecContext(ctx, query,
		workspace.ID,
		workspace.Repo,
		workspace.Owner,
		workspace.Ref,
		workspace.DesiredState,
		workspace.ActualState,
		workspace.ExecutionProfile,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)
	return err
}

func (r *WorkspaceRepository) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	const query = `SELECT id, repo, owner, ref, desired_state, actual_state, execution_profile, created_at, updated_at FROM workspaces WHERE id = $1`

	workspace := &domain.Workspace{}
	var executionProfile []byte
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&workspace.ID,
		&workspace.Repo,
		&workspace.Owner,
		&workspace.Ref,
		&workspace.DesiredState,
		&workspace.ActualState,
		&executionProfile,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrWorkspaceNotFound
		}
		return nil, err
	}
	workspace.ExecutionProfile = executionProfile
	return workspace, nil
}

func (r *WorkspaceRepository) List(ctx context.Context) ([]*domain.Workspace, error) {
	const query = `SELECT id, repo, owner, ref, desired_state, actual_state, execution_profile, created_at, updated_at FROM workspaces ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.Workspace
	for rows.Next() {
		workspace := &domain.Workspace{}
		var executionProfile []byte
		if err := rows.Scan(
			&workspace.ID,
			&workspace.Repo,
			&workspace.Owner,
			&workspace.Ref,
			&workspace.DesiredState,
			&workspace.ActualState,
			&executionProfile,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		); err != nil {
			return nil, err
		}
		workspace.ExecutionProfile = executionProfile
		result = append(result, workspace)
	}

	return result, rows.Err()
}

func (r *WorkspaceRepository) UpdateDesiredState(ctx context.Context, id, desiredState string, updatedAt time.Time) error {
	const query = `UPDATE workspaces SET desired_state = $1, updated_at = $2 WHERE id = $3`
	res, err := r.db.ExecContext(ctx, query, desiredState, updatedAt, id)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err != nil {
		return err
	} else if affected == 0 {
		return domain.ErrWorkspaceNotFound
	}
	return nil
}

func (r *WorkspaceRepository) UpdateActualState(ctx context.Context, id, actualState string, updatedAt time.Time) error {
	const query = `UPDATE workspaces SET actual_state = $1, updated_at = $2 WHERE id = $3`
	res, err := r.db.ExecContext(ctx, query, actualState, updatedAt, id)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err != nil {
		return err
	} else if affected == 0 {
		return domain.ErrWorkspaceNotFound
	}
	return nil
}
