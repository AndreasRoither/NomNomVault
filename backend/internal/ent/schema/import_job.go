package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ImportJob stores one household-scoped import attempt and its review metadata.
type ImportJob struct {
	ent.Schema
}

func (ImportJob) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}, HouseholdScopedMixin{}}
}

func (ImportJob) Fields() []ent.Field {
	return []ent.Field{
		field.String("requested_by_user_id"),
		field.String("source_record_id"),
		field.Enum("import_kind").Values(ImportKindValues...),
		field.Enum("status").Values(ImportJobStatusValues...).Default("queued"),
		field.String("idempotency_key").Optional().Nillable(),
		field.String("active_idempotency_key").Optional().Nillable().Unique(),
		field.String("active_fingerprint_key").Optional().Nillable().Unique(),
		field.String("fallback_fingerprint").NotEmpty(),
		field.JSON("normalized_payload_json", map[string]any{}).Optional(),
		field.String("draft_recipe_id").Optional().Nillable(),
		field.String("match_recipe_id").Optional().Nillable(),
		field.Enum("conflict_state").Values(ConflictStateValues...).Default("none"),
		field.JSON("warnings_json", []string{}).Optional(),
		field.Float("confidence_score").Optional().Nillable(),
		field.String("error_code").Optional().Nillable(),
		field.String("error_message").Optional().Nillable(),
		field.Int("attempt_count").Default(1),
		field.Time("started_at").Optional().Nillable(),
		field.Time("finished_at").Optional().Nillable(),
	}
}

func (ImportJob) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("household", Household.Type).Ref("import_jobs").Field("household_id").Unique().Required(),
		edge.From("requested_by_user", User.Type).Ref("requested_import_jobs").Field("requested_by_user_id").Unique().Required(),
		edge.From("source_record", SourceRecord.Type).
			Ref("import_jobs").
			Field("source_record_id").
			Unique().
			Required(),
		edge.From("draft_recipe", Recipe.Type).Ref("draft_import_jobs").Field("draft_recipe_id").Unique(),
		edge.From("match_recipe", Recipe.Type).Ref("matched_import_jobs").Field("match_recipe_id").Unique(),
	}
}

func (ImportJob) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("household_id", "requested_by_user_id", "import_kind", "idempotency_key"),
		index.Fields("household_id", "import_kind", "fallback_fingerprint"),
		index.Fields("household_id", "status", "created_at"),
		index.Fields("source_record_id", "created_at"),
	}
}
