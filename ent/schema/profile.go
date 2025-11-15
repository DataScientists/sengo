package schema

import (
	"sheng-go-backend/ent/mixin"
	"sheng-go-backend/pkg/const/globalid"

	entMixin "entgo.io/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
		// LinkedIn URN - unique identifier
		field.String("urn").
			NotEmpty().
			Unique().
			Comment("LinkedIn profile URN identifier"),

		// Basic profile info
		field.String("username").
			Optional().
			Nillable().
			Comment("LinkedIn username from URL"),

		field.String("first_name").
			Optional().
			Nillable().
			Comment("First name"),

		field.String("last_name").
			Optional().
			Nillable().
			Comment("Last name"),

		field.String("name").
			Optional().
			Nillable().
			MaxLen(255).
			Comment("Full name (legacy field)"),

		field.String("headline").
			Optional().
			Nillable().
			Comment("LinkedIn headline/bio"),

		field.String("title").
			Optional().
			Nillable().
			Comment("Current job title (legacy field)"),

		// Geographic info
		field.String("country").
			Optional().
			Nillable().
			Comment("Country name"),

		field.String("city").
			Optional().
			Nillable().
			Comment("City name"),

		// JSON arrays for nested data
		field.JSON("educations", []map[string]interface{}{}).
			Optional().
			Comment("Array of education records: [{schoolName, degree?, fieldOfStudy?}]"),

		field.JSON("positions", []map[string]interface{}{}).
			Optional().
			Comment("Array of position records: [{companyName, title, description?, isCurrent?}]"),

		field.JSON("skills", []map[string]interface{}{}).
			Optional().
			Comment("Array of skills: [{name, endorsementCount?}]"),

		field.JSON("geo_data", map[string]interface{}{}).
			Optional().
			Comment("Geographic data: {country_name, city_name}"),

		// S3 storage keys
		field.String("raw_data_s3_key").
			Optional().
			Nillable().
			MaxLen(500).
			Comment("S3 path to full RapidAPI response JSON"),

		field.String("cleaned_data_s3_key").
			Optional().
			Nillable().
			MaxLen(500).
			Comment("S3 path to cleaned/extracted profile JSON"),

		field.String("source_file").
			Optional().
			Nillable().
			Comment("Legacy source file field"),
	}
}

// Edges of the Profile.
func (Profile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("profile_entry", ProfileEntry.Type).
			Ref("profile").
			Unique().
			Comment("Associated ProfileEntry that triggered this profile fetch"),
	}
}

// Indexes of the Profile.
func (Profile) Indexes() []ent.Index {
	return []ent.Index{
		// Index for searching by name
		index.Fields("first_name"),
		index.Fields("last_name"),
		index.Fields("username"),

		// Index for location-based queries
		index.Fields("country"),
		index.Fields("city"),
		index.Fields("country", "city"),

		// Index for URN lookups (already unique, but explicit)
		index.Fields("urn"),
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
