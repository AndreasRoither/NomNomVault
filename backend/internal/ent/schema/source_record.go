package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SourceRecord stores import provenance and retained source artifacts.
type SourceRecord struct {
	ent.Schema
}

func (SourceRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}, HouseholdScopedMixin{}}
}

func (SourceRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("source_type").Values(SourceTypeValues...),
		field.Enum("import_kind").Values(ImportKindValues...),
		field.String("submitted_url").Optional().Nillable(),
		field.String("normalized_url").Optional().Nillable(),
		field.String("canonical_url").Optional().Nillable(),
		field.String("title_hint").Optional().Nillable(),
		field.String("content_hash").Optional().Nillable(),
		field.String("raw_snapshot_storage_object_id").Optional().Nillable(),
		field.JSON("metadata_json", map[string]any{}).Optional(),
		field.Enum("retention_state").Values(SourceRetentionStateValues...).Default("retained"),
	}
}

func (SourceRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("household", Household.Type).Ref("source_records").Field("household_id").Unique().Required(),
		edge.From("raw_snapshot_storage_object", StoredObject.Type).
			Ref("source_records").
			Field("raw_snapshot_storage_object_id").
			Unique(),
		edge.To("import_jobs", ImportJob.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (SourceRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("household_id", "import_kind", "normalized_url"),
		index.Fields("household_id", "import_kind", "content_hash"),
	}
}
