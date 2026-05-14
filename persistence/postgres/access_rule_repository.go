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
	const updateQuery = `UPDATE access_rules SET role = $3 WHERE "type" = $1 AND value = $2`
	res, err := r.db.ExecContext(ctx, updateQuery, subjectType, subjectID, role)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected > 0 {
		return nil
	}

	const insertQuery = `INSERT INTO access_rules (id, "type", value, role, created_at) VALUES (gen_random_uuid(), $1, $2, $3, now())`
	_, err = r.db.ExecContext(ctx, insertQuery, subjectType, subjectID, role)
	return err
}

func (r *AccessRuleRepository) Delete(ctx context.Context, subjectType, subjectID string) error {
	const query = `DELETE FROM access_rules WHERE "type" = $1 AND value = $2`
	_, err := r.db.ExecContext(ctx, query, subjectType, subjectID)
	return err
}
