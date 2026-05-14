package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AccessRule holds the schema definition for the AccessRule entity.
type AccessRule struct {
	ent.Schema
}

// Fields of the AccessRule.
func (AccessRule) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		field.String("type").NotEmpty(),
		field.String("value").NotEmpty(),
		field.String("role").NotEmpty(),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the AccessRule.
func (AccessRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type", "value").Unique(),
	}
}

func (AccessRule) Edges() []ent.Edge {
	return nil
}
