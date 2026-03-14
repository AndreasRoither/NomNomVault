package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// HouseholdScopedMixin adds explicit household ownership without request hooks.
type HouseholdScopedMixin struct{}

func (HouseholdScopedMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("household_id"),
	}
}

func (HouseholdScopedMixin) Edges() []ent.Edge                { return nil }
func (HouseholdScopedMixin) Annotations() []schema.Annotation { return nil }
func (HouseholdScopedMixin) Hooks() []ent.Hook                { return nil }
func (HouseholdScopedMixin) Indexes() []ent.Index             { return nil }
func (HouseholdScopedMixin) Interceptors() []ent.Interceptor  { return nil }
func (HouseholdScopedMixin) Policy() ent.Policy               { return nil }
