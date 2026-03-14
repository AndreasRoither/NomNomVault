package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Tag for recipe categorization.
type Tag struct {
	ent.Schema
}

func (Tag) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}, HouseholdScopedMixin{}}
}

func (Tag) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("slug").NotEmpty(),
		field.String("color").Default(""),
		field.Bool("system").Default(false),
	}
}

func (Tag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("household", Household.Type).Ref("tags").Field("household_id").Unique().Required(),
		edge.From("recipes", Recipe.Type).Ref("tags"),
	}
}

func (Tag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("household_id", "slug").Unique(),
	}
}
