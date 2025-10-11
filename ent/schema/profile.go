package schema

import (
	"sheng-go-backend/ent/mixin"
	"sheng-go-backend/pkg/const/globalid"

	entMixin "entgo.io/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// holds the schema definition for the Test entity.
type Profile struct {
	ent.Schema
}

// ProfileMixin defines Fields
type ProfileMixin struct {
	entMixin.Schema
}

// Fields of the Profile.
func (ProfileMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(255),
		field.String("title"),
		field.String("urn").NotEmpty().Unique(),
		field.String("source_file").NotEmpty().Unique(),
	}
}

// Edges of the Profile.
func (Profile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("todos", Todo.Type),
	}
}

// Mixin of the Profile.
func (Profile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewUlid(globalid.New().Profile.Prefix),
		ProfileMixin{},
		mixin.NewDatetime(),
	}
}
