package schema

import (
	"sheng-go-backend/ent/mixin"
	"sheng-go-backend/pkg/const/globalid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	entMixin "entgo.io/ent/schema/mixin"
)

// ProfilePostItem stores a single LinkedIn post fetched for a profile.
type ProfilePostItem struct {
	ent.Schema
}

type ProfilePostItemMixin struct {
	entMixin.Schema
}

func (ProfilePostItemMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("profile_username").
			NotEmpty().
			Comment("LinkedIn username this post belongs to"),

		field.String("post_urn").
			Optional().
			Nillable().
			Comment("LinkedIn post URN (numeric ID string from API)"),

		field.String("post_url").
			Optional().
			Nillable().
			Comment("postUrl from API response"),

		field.Text("text").
			Optional().
			Nillable().
			Comment("Post body text"),

		field.String("content_type").
			Optional().
			Nillable().
			Comment("contentType: post, image, article, linkedInVideo, etc."),

		field.Bool("is_repost").
			Default(false).
			Comment("Whether this is a reshared post"),

		field.Int("total_reactions").
			Default(0).
			NonNegative().
			Comment("totalReactionCount from API"),

		field.Int("like_count").
			Default(0).
			NonNegative().
			Comment("likeCount from API"),

		field.Int("comments_count").
			Default(0).
			NonNegative().
			Comment("commentsCount from API"),

		field.Int("reposts_count").
			Default(0).
			NonNegative().
			Comment("repostsCount from API"),

		field.Int("empathy_count").
			Default(0).
			NonNegative().
			Comment("empathyCount (care reactions) from API"),

		field.Int("praise_count").
			Default(0).
			NonNegative().
			Comment("praiseCount (celebrate reactions) from API"),

		field.Int("funny_count").
			Default(0).
			NonNegative().
			Comment("funnyCount reactions from API"),

		field.Int("interest_count").
			Default(0).
			NonNegative().
			Comment("InterestCount (insightful reactions) from API"),

		field.Time("posted_at").
			Optional().
			Nillable().
			Comment("Parsed from postedDateTimestamp (ms unix)"),

		field.JSON("raw_data", map[string]interface{}{}).
			Optional().
			Comment("Full raw post object from RapidAPI"),
	}
}

func (ProfilePostItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewUlid(globalid.New().ProfilePostItem.Prefix),
		ProfilePostItemMixin{},
		mixin.NewDatetime(),
	}
}

func (ProfilePostItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("profile_post", ProfilePost.Type).
			Ref("items").
			Unique().
			Comment("Parent ProfilePost record"),
	}
}

func (ProfilePostItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("profile_username"),
		index.Fields("profile_username", "posted_at"),
	}
}
