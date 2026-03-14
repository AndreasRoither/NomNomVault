package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RefreshSession stores a server-side refresh record for cookie-based auth.
type RefreshSession struct {
	ent.Schema
}

func (RefreshSession) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (RefreshSession) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id"),
		field.String("token_hash").NotEmpty().Unique(),
		field.Time("expires_at"),
		field.Bool("revoked").Default(false),
		field.String("device_info").Optional().Nillable(),
		field.String("ip_address").Optional().Nillable(),
		field.Time("last_used_at").Optional().Nillable(),
	}
}

func (RefreshSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("refresh_sessions").Field("user_id").Unique().Required(),
	}
}

func (RefreshSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "revoked"),
		index.Fields("expires_at"),
	}
}
