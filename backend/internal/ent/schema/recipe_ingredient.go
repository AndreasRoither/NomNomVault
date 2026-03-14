package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RecipeIngredient stores one ingredient row for a recipe.
type RecipeIngredient struct {
	ent.Schema
}

func (RecipeIngredient) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (RecipeIngredient) Fields() []ent.Field {
	return []ent.Field{
		field.String("recipe_id"),
		field.String("name").NotEmpty(),
		field.Float("quantity").Optional().Nillable(),
		field.Enum("unit").Values(UnitValues...).Optional().Nillable(),
		field.String("preparation").Optional().Nillable(),
		field.Int("sort_order"),
	}
}

func (RecipeIngredient) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("recipe", Recipe.Type).Ref("ingredients").Field("recipe_id").Unique().Required(),
	}
}

func (RecipeIngredient) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("recipe_id", "sort_order"),
	}
}
