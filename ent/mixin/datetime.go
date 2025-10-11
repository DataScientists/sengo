package mixin

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// NewDatetime creates a Mixin that includes create_at and updated_at
func NewDatetime() *DatetimeMixin {
	return &DatetimeMixin{}
}

// DatetimeMixin defines an ent Mixin
type DatetimeMixin struct {
	mixin.Schema
}

// Fields provides the created_at and updated_at field.
func (m DatetimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz",
			}).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			// Ensure this field is updated automatically at the application level
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz",
			}).Annotations(
			entgql.Skip(entgql.SkipWhereInput), // Skip all inputs
		),
	}
}
