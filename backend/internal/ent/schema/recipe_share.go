package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RecipeShare stores hashed public share links for anonymous recipe viewing.
type RecipeShare struct {
	ent.Schema
}

func (RecipeShare) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (RecipeShare) Fields() []ent.Field {
	return []ent.Field{
		field.String("recipe_id"),
		field.String("token_hash").NotEmpty().Unique().Immutable(),
		field.String("created_by_user_id"),
		field.Time("expires_at").Optional().Nillable(),
		field.Time("revoked_at").Optional().Nillable(),
		field.Time("last_accessed_at").Optional().Nillable(),
		field.Int("access_count").Default(0),
	}
}

func (RecipeShare) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("recipe", Recipe.Type).Ref("shares").Field("recipe_id").Unique().Required(),
		edge.From("created_by", User.Type).Ref("recipe_shares").Field("created_by_user_id").Unique().Required(),
	}
}

func (RecipeShare) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("recipe_id"),
		index.Fields("created_by_user_id"),
		index.Fields("expires_at"),
	}
}
