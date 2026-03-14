package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// HouseholdMember links a user to a household with a specific role.
type HouseholdMember struct {
	ent.Schema
}

func (HouseholdMember) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}, HouseholdScopedMixin{}}
}

func (HouseholdMember) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id"),
		field.Enum("role").Values("viewer", "editor", "owner").Default("viewer"),
	}
}

func (HouseholdMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("household", Household.Type).Ref("members").Field("household_id").Unique().Required(),
		edge.From("user", User.Type).Ref("memberships").Field("user_id").Unique().Required(),
	}
}

func (HouseholdMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("household_id", "user_id").Unique(),
	}
}
