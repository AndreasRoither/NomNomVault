package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RecipeNutrition stores structured nutrition data for a recipe.
type RecipeNutrition struct {
	ent.Schema
}

func (RecipeNutrition) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (RecipeNutrition) Fields() []ent.Field {
	return []ent.Field{
		field.String("recipe_id"),
		field.String("reference_quantity").Optional().Nillable(),
		field.Int("energy_kj").Optional().Nillable(),
		field.Int("energy_kcal").Optional().Nillable(),
		field.Float("protein").Optional().Nillable(),
		field.Float("carbohydrates").Optional().Nillable(),
		field.Float("fat").Optional().Nillable(),
		field.Float("saturated_fat").Optional().Nillable(),
		field.Float("fiber").Optional().Nillable(),
		field.Float("sugars").Optional().Nillable(),
		field.Float("sodium").Optional().Nillable(),
		field.Float("salt").Optional().Nillable(),
		field.JSON("breakdown", map[string]any{}).Optional(),
	}
}

func (RecipeNutrition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("recipe", Recipe.Type).Ref("nutrition_entries").Field("recipe_id").Unique().Required(),
	}
}

func (RecipeNutrition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("recipe_id", "reference_quantity").Unique(),
	}
}

func (RecipeNutrition) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "recipe_nutrition"},
	}
}
