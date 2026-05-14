package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Workspace holds the schema definition for the Workspace entity.
type Workspace struct {
	ent.Schema
}

// Fields of the Workspace.
func (Workspace) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").NotEmpty().Unique(),
		field.String("repo").NotEmpty(),
		field.String("owner").NotEmpty(),
		field.String("ref").NotEmpty(),
		field.String("desired_state").NotEmpty(),
		field.String("actual_state").NotEmpty(),
		field.JSON("execution_profile", json.RawMessage{}).Default(json.RawMessage([]byte("{}"))),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Workspace.
func (Workspace) Edges() []ent.Edge {
	return nil
}
