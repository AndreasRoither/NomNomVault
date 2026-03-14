package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RecipeStep stores one ordered instruction step for a recipe.
type RecipeStep struct {
	ent.Schema
}

func (RecipeStep) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (RecipeStep) Fields() []ent.Field {
	return []ent.Field{
		field.String("recipe_id"),
		field.Int("sort_order"),
		field.String("instruction").NotEmpty(),
		field.Int("duration_minutes").Optional().Nillable(),
		field.String("tip").Optional().Nillable(),
	}
}

func (RecipeStep) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("recipe", Recipe.Type).Ref("steps").Field("recipe_id").Unique().Required(),
	}
}

func (RecipeStep) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("recipe_id", "sort_order"),
	}
}
