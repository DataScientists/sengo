package schema

import (
	"sheng-go-backend/ent/mixin"
	"sheng-go-backend/pkg/const/globalid"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	entMixin "entgo.io/ent/schema/mixin"
)

// ProfilePost tracks the post-fetch status for a single LinkedIn profile.
// One row per profile; individual posts are stored in ProfilePostItem.
type ProfilePost struct {
	ent.Schema
}

type ProfilePostMixin struct {
	entMixin.Schema
}

func (ProfilePostMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("profile_username").
			NotEmpty().
			Unique().
			Comment("LinkedIn username extracted from profile_url"),

		field.Enum("fetch_status").
			NamedValues(
				"Pending", "PENDING",
				"COMPLETED", "COMPLETED",
				"FAILED", "FAILED",
				"NotFound", "NOT_FOUND",
			).
			Annotations(entgql.Type("ProfilePostFetchStatus")).
			Default("PENDING").
			Comment("Fetch status for this profile's posts"),

		field.String("s3_key").
			Optional().
			Nillable().
			MaxLen(500).
			Comment("S3 key for the raw posts JSON file"),

		field.Text("error_message").
			Optional().
			Nillable().
			Comment("Error message if fetch failed"),
	}
}

func (ProfilePost) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewUlid(globalid.New().ProfilePost.Prefix),
		ProfilePostMixin{},
		mixin.NewDatetime(),
	}
}

func (ProfilePost) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("items", ProfilePostItem.Type).
			Comment("Individual post items fetched for this profile"),
	}
}

func (ProfilePost) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("profile_username"),
		index.Fields("fetch_status"),
		index.Fields("fetch_status", "created_at"),
	}
}
