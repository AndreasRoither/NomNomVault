package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// MediaAsset holds the schema definition for the MediaAsset entity.
type MediaAsset struct {
	ent.Schema
}

func (MediaAsset) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}, HouseholdScopedMixin{}}
}

func (MediaAsset) Fields() []ent.Field {
	return []ent.Field{
		field.String("recipe_id").Optional().Nillable(),
		field.String("storage_object_id"),
		field.String("original_filename").NotEmpty(),
		field.String("mime_type").NotEmpty(),
		field.Enum("media_type").Values("image", "document", "other").Default("image"),
		field.Int64("size_bytes"),
		field.String("checksum").NotEmpty(),
		field.Time("stored_at"),
		field.String("alt_text").Default(""),
		field.Int("sort_order").Default(1),
	}
}

func (MediaAsset) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("household", Household.Type).Ref("media_assets").Field("household_id").Unique().Required(),
		edge.From("recipe", Recipe.Type).Ref("media_assets").Field("recipe_id").Unique(),
		edge.From("storage_object", StoredObject.Type).Ref("media_assets").Field("storage_object_id").Unique().Required(),
	}
}

func (MediaAsset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("recipe_id", "sort_order"),
	}
}
