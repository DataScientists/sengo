package schema

import (
	"sheng-go-backend/ent/mixin"
	"sheng-go-backend/ent/schema/ulid"
	"sheng-go-backend/pkg/const/globalid"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	entMixin "entgo.io/ent/schema/mixin"
)

// Todo holds the schema definition for the Todo entity.
type Todo struct {
	ent.Schema
}

// TodoMixin defines Fields
type TodoMixin struct {
	entMixin.Schema
}

// Fields of the Todo.
func (TodoMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").
			GoType(ulid.ID("")).
			Optional(),
		field.String("name").Default(""),
		field.Enum("status").
			NamedValues(
				"InProgress", "IN_PROGRESS",
				"Completed", "COMPLETED",
			).Annotations(
			entgql.Type("TodoStatus"), // Customize the enum name
		).
			Default("IN_PROGRESS"),
		field.Int("priority").Default(0),
	}
}

// Edges of the Todo.
func (Todo) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("todos").
			Unique().
			Field("user_id"),
	}
}

// Mixin of the Todo.
func (Todo) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewUlid(globalid.New().Todo.Prefix),
		TodoMixin{},
		mixin.NewDatetime(),
	}
}
