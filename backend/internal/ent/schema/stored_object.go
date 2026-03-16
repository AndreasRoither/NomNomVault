package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// StoredObject stores immutable household-scoped blob content in PostgreSQL.
// For the links to recipes, see the `media_assets` edge.
type StoredObject struct {
	ent.Schema
}

func (StoredObject) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}, HouseholdScopedMixin{}}
}

func (StoredObject) Fields() []ent.Field {
	return []ent.Field{
		field.String("original_filename").NotEmpty(),
		field.String("mime_type").NotEmpty(),
		field.Int64("size_bytes"),
		field.String("checksum").NotEmpty(),
		field.Bytes("content"),
	}
}

func (StoredObject) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("household", Household.Type).Ref("stored_objects").Field("household_id").Unique().Required(),
		edge.To("media_assets", MediaAsset.Type),
		edge.To("thumbnail_media_assets", MediaAsset.Type),
		edge.To("source_records", SourceRecord.Type),
	}
}

func (StoredObject) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("household_id", "checksum").Unique(),
	}
}
