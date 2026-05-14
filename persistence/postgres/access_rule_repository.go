package postgres

import (
	"context"
	"database/sql"

	"github.com/UtopikCode/quickspaces-control-plane/internal/application/auth"
)

type AccessRuleRepository struct {
	db *sql.DB
}

func NewAccessRuleRepository(db *sql.DB) *AccessRuleRepository {
	return &AccessRuleRepository{db: db}
}

func (r *AccessRuleRepository) List(ctx context.Context) ([]*auth.AccessRule, error) {
	const query = `SELECT "type", value, role FROM access_rules`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	rules := make([]*auth.AccessRule, 0)
	for rows.Next() {
		rule := &auth.AccessRule{}
		if err := rows.Scan(&rule.SubjectType, &rule.SubjectID, &rule.Role); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

func (r *AccessRuleRepository) Upsert(ctx context.Context, subjectType, subjectID, role string) error {
	const query = `INSERT INTO access_rules ("type", value, role) VALUES ($1, $2, $3) ON CONFLICT ("type", value) DO UPDATE SET role = EXCLUDED.role`
	_, err := r.db.ExecContext(ctx, query, subjectType, subjectID, role)
	return err
}

func (r *AccessRuleRepository) Delete(ctx context.Context, subjectType, subjectID string) error {
	const query = `DELETE FROM access_rules WHERE "type" = $1 AND value = $2`
	_, err := r.db.ExecContext(ctx, query, subjectType, subjectID)
	return err
}
