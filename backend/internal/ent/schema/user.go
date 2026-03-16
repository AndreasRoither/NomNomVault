package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("display_name").NotEmpty(),
		field.String("email").NotEmpty().Unique(),
		field.Bool("email_verified").Default(false),
		field.String("password_hash").NotEmpty(),
		field.Enum("role").Values("user", "sys_admin").Default("user"),
		field.Time("last_login_at").Optional().Nillable(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("memberships", HouseholdMember.Type),
		edge.To("recipe_shares", RecipeShare.Type),
		edge.To("requested_import_jobs", ImportJob.Type),
		edge.To("refresh_sessions", RefreshSession.Type),
	}
}
