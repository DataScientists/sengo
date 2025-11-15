package schema

import (
	"sheng-go-backend/ent/mixin"
	"sheng-go-backend/pkg/const/globalid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	entMixin "entgo.io/ent/schema/mixin"
)

// APIQuotaTracker holds the schema definition for API quota tracking
type APIQuotaTracker struct {
	ent.Schema
}

// APIQuotaTrackerMixin defines Fields
type APIQuotaTrackerMixin struct {
	entMixin.Schema
}

// Fields of the APIQuotaTracker.
func (APIQuotaTracker) Fields() []ent.Field {
	return []ent.Field{
		// Time period
		field.Int("month").
			Range(1, 12).
			Comment("Month (1-12)"),

		field.Int("year").
			Positive().
			Comment("Year"),

		// Quota tracking
		field.Int("call_count").
			Default(0).
			NonNegative().
			Comment("Number of API calls made this month"),

		field.Int("quota_limit").
			Default(50000).
			Positive().
			Comment("Monthly API call limit"),

		field.Bool("quota_exceeded").
			Default(false).
			Comment("Whether quota has been exceeded"),

		field.Bool("override_enabled").
			Default(false).
			Comment("Admin override to bypass quota limits"),

		field.Bool("notification_sent").
			Default(false).
			Comment("Whether quota exceeded notification has been sent"),

		field.Time("last_call_at").
			Optional().
			Nillable().
			Comment("Timestamp of last API call"),
	}
}

// Mixin of the APIQuotaTracker.
func (APIQuotaTracker) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewUlid(globalid.New().APIQuotaTracker.Prefix),
		APIQuotaTrackerMixin{},
		mixin.NewDatetime(),
	}
}

// Indexes of the APIQuotaTracker.
func (APIQuotaTracker) Indexes() []ent.Index {
	return []ent.Index{
		// Unique constraint on month + year
		index.Fields("month", "year").Unique(),

		// Index for finding current/recent trackers
		index.Fields("year", "month"),
	}
}
