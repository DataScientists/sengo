package schema

import (
	"sheng-go-backend/ent/mixin"
	"sheng-go-backend/pkg/const/globalid"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	entMixin "entgo.io/ent/schema/mixin"
)

// ProfileEntry holds the schema definition for the ProfileEntry entity.
type ProfileEntry struct {
	ent.Schema
}

// ProfileMixin defines Fields
type ProfileEntryMixin struct {
	entMixin.Schema
}

// Fields of the ProfileEntry.
func (ProfileEntry) Fields() []ent.Field {
	return []ent.Field{
		field.String("linkedin_urn").
			NotEmpty().
			Unique().
			Comment("LinkedIn profile URN identifier"),

		field.String("gender").
			Optional().
			Nillable().
			Comment("User-provided gender information"),

		field.Enum("status").
			NamedValues(
				"Pending", "PENDING",
				"Fetching", "FETCHING",
				"COMPLETED", "COMPLETED",
				"FAILED", "FAILED",
			).Annotations(
			entgql.Type("ProfileEntryStatus")).
			Default("PENDING").
			Comment("Current fetch status of the profile"),

		field.JSON("profile_data", map[string]any{}).
			Optional().
			Comment("Extracted profile data based on template"),

		field.String("template_json_s3_key").
			Optional().
			Nillable().
			MaxLen(500).
			Comment("S3 path to template JSON file"),

		field.String("raw_response_s3_key").
			Optional().
			Nillable().
			MaxLen(500).
			Comment("S3 path to full API response JSON"),

		field.Int("fetch_count").
			Default(0).
			NonNegative().
			Comment("Number of times this profile has been fetched"),

		field.Time("last_fetched_at").
			Optional().
			Nillable().
			Comment("Timestamp of last successful fetch"),

		field.Text("error_message").
			Optional().
			Nillable().
			Comment("Error message if fetch failed"),
	}
}

func (ProfileEntry) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewUlid(globalid.New().ProfileEntry.Prefix),
		ProfileEntryMixin{},
		mixin.NewDatetime(),
	}
}

// Indexes of the ProfileEntry.
func (ProfileEntry) Indexes() []ent.Index {
	return []ent.Index{
		// Index for filtering by status
		index.Fields("status"),

		// Index for searching by URN (already unique, but explicit index)
		index.Fields("linkedin_urn"),

		// Index for sorting by creation date (descending order most common)
		index.Fields("created_at"),

		// Composite index for status + created_at queries
		// Useful for "get pending profiles ordered by creation"
		index.Fields("status", "created_at"),
	}
}

// Edges of the ProfileEntry.
func (ProfileEntry) Edges() []ent.Edge {
	return []ent.Edge{
		// Add edges here when you add related entities
		// For example, when you add fetch_jobs:
		// edge.From("fetch_jobs", FetchJob.Type).Ref("profile_entries"),
	}
}
