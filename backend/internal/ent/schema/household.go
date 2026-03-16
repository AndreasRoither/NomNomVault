package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Household holds the schema definition for the Household entity.
type Household struct {
	ent.Schema
}

func (Household) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Household) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("slug").NotEmpty().Unique(),
		field.String("description").Default(""),
		field.Bool("active").Default(true),
	}
}

func (Household) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("members", HouseholdMember.Type),
		edge.To("recipes", Recipe.Type),
		edge.To("tags", Tag.Type),
		edge.To("media_assets", MediaAsset.Type),
		edge.To("stored_objects", StoredObject.Type),
		edge.To("refresh_sessions", RefreshSession.Type),
	}
}
