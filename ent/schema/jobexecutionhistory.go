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

// JobExecutionHistory holds the schema definition for tracking job execution history
type JobExecutionHistory struct {
	ent.Schema
}

// JobExecutionHistoryMixin defines Fields
type JobExecutionHistoryMixin struct {
	entMixin.Schema
}

// Fields of the JobExecutionHistory.
func (JobExecutionHistory) Fields() []ent.Field {
	return []ent.Field{
		// Job identification
		field.String("job_name").
			NotEmpty().
			MaxLen(100).
			Comment("Name of the executed job"),

		field.Enum("status").
			NamedValues(
				"Success", "SUCCESS",
				"Failed", "FAILED",
				"Partial", "PARTIAL",
				"QuotaExceeded", "QUOTA_EXCEEDED",
			).
			Annotations(entgql.Type("JobExecutionStatus")).
			Comment("Execution status"),

		// Timing
		field.Time("started_at").
			Comment("Job start time"),

		field.Time("completed_at").
			Optional().
			Nillable().
			Comment("Job completion time"),

		field.Int("duration_seconds").
			Default(0).
			NonNegative().
			Comment("Duration in seconds"),

		// Statistics
		field.Int("total_processed").
			Default(0).
			NonNegative().
			Comment("Total items attempted"),

		field.Int("successful_count").
			Default(0).
			NonNegative().
			Comment("Successfully processed items"),

		field.Int("failed_count").
			Default(0).
			NonNegative().
			Comment("Failed items"),

		field.Int("api_calls_made").
			Default(0).
			NonNegative().
			Comment("Number of API calls made during this execution"),

		field.Int("quota_remaining").
			Default(0).
			NonNegative().
			Comment("API quota remaining after execution"),

		// Error tracking
		field.Text("error_summary").
			Optional().
			Nillable().
			Comment("Summary of errors encountered"),
	}
}

// Mixin of the JobExecutionHistory.
func (JobExecutionHistory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewUlid(globalid.New().JobExecutionHistory.Prefix),
		JobExecutionHistoryMixin{},
		mixin.NewDatetime(),
	}
}

// Indexes of the JobExecutionHistory.
func (JobExecutionHistory) Indexes() []ent.Index {
	return []ent.Index{
		// Index for filtering by job name
		index.Fields("job_name"),

		// Index for filtering by status
		index.Fields("status"),

		// Index for time-based queries (most recent first)
		index.Fields("started_at").
			StorageKey("idx_started_at_desc"),

		// Composite index for job + time
		index.Fields("job_name", "started_at"),
	}
}
