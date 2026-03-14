package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// BaseMixin provides the shared string ID and timestamps used by all entities.
type BaseMixin struct{}

func (BaseMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable().DefaultFunc(uuid.NewString),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (BaseMixin) Edges() []ent.Edge                { return nil }
func (BaseMixin) Annotations() []schema.Annotation { return nil }
func (BaseMixin) Hooks() []ent.Hook                { return nil }
func (BaseMixin) Indexes() []ent.Index             { return nil }
func (BaseMixin) Interceptors() []ent.Interceptor  { return nil }
func (BaseMixin) Policy() ent.Policy               { return nil }
