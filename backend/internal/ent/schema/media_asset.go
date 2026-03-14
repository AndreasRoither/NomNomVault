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
		field.String("original_filename").NotEmpty(),
		field.String("mime_type").NotEmpty(),
		field.Enum("media_type").Values("image", "document", "other").Default("image"),
		field.Int64("size_bytes"),
		field.String("checksum").NotEmpty(),
		field.Time("stored_at"),
	}
}

func (MediaAsset) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("household", Household.Type).Ref("media_assets").Field("household_id").Unique().Required(),
		edge.From("recipe", Recipe.Type).Ref("media_assets").Field("recipe_id").Unique(),
	}
}

func (MediaAsset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("household_id", "checksum").Unique(),
	}
}
