package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Recipe holds the schema definition for the Recipe entity.
type Recipe struct {
	ent.Schema
}

func (Recipe) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}, HouseholdScopedMixin{}}
}

func (Recipe) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty(),
		field.String("description").Default(""),
		field.Enum("status").Values(RecipeStatusValues...).Default("published"),
		field.String("source_url").Default(""),
		field.Time("source_captured_at").Optional().Nillable(),
		field.String("primary_media_id").Optional().Nillable(),
		field.JSON("gallery_media_ids", []string{}).Optional(),
		field.Int("prep_minutes").Optional().Nillable(),
		field.Int("cook_minutes").Optional().Nillable(),
		field.Int("servings").Optional().Nillable(),
		field.Enum("region").Values(RegionValues...).Optional().Nillable(),
		field.Enum("meal_type").Values(MealTypeValues...).Optional().Nillable(),
		field.Enum("difficulty").Values(DifficultyValues...).Optional().Nillable(),
		field.Enum("cuisine").Values(CuisineValues...).Optional().Nillable(),
		field.Float("popularity_score").Optional().Nillable().Default(0),
		field.Strings("allergens").Optional().Validate(ValidateAllergenValues),
		field.Time("aggregated_at").Optional().Nillable(),
		field.Int("version").Default(1),
	}
}

func (Recipe) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("household", Household.Type).Ref("recipes").Field("household_id").Unique().Required(),
		edge.To("ingredients", RecipeIngredient.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("steps", RecipeStep.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("nutrition_entries", RecipeNutrition.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("shares", RecipeShare.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("tags", Tag.Type),
		edge.To("media_assets", MediaAsset.Type).Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("draft_import_jobs", ImportJob.Type),
		edge.To("matched_import_jobs", ImportJob.Type),
	}
}

func (Recipe) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("household_id", "title"),
	}
}
