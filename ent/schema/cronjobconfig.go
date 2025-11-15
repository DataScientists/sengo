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

// CronJobConfig holds the schema definition for cron job configuration
type CronJobConfig struct {
	ent.Schema
}

// CronJobConfigMixin defines Fields
type CronJobConfigMixin struct {
	entMixin.Schema
}

// Fields of the CronJobConfig.
func (CronJobConfig) Fields() []ent.Field {
	return []ent.Field{
		// Job identification
		field.String("job_name").
			NotEmpty().
			Unique().
			MaxLen(100).
			Comment("Unique name for the cron job"),

		field.Enum("job_type").
			NamedValues(
				"ProfileFetcher", "PROFILE_FETCHER",
				"QuotaReset", "QUOTA_RESET",
			).
			Annotations(entgql.Type("CronJobType")).
			Comment("Type of cron job"),

		// Schedule configuration
		field.String("schedule").
			NotEmpty().
			Comment("Cron expression (e.g., '0 2 * * *' for 2 AM daily)"),

		field.Bool("enabled").
			Default(true).
			Comment("Whether the job is enabled"),

		// Job parameters
		field.Int("batch_size").
			Default(10).
			Positive().
			Comment("Number of items to process per job run"),

		field.String("admin_email").
			NotEmpty().
			Comment("Admin email for notifications"),

		field.Bool("respect_quota").
			Default(true).
			Comment("Whether this job should respect API quota limits"),

		// Execution tracking
		field.Time("last_run_at").
			Optional().
			Nillable().
			Comment("Timestamp of last successful run"),

		field.Time("next_run_at").
			Optional().
			Nillable().
			Comment("Scheduled next run time"),
	}
}

// Mixin of the CronJobConfig.
func (CronJobConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewUlid(globalid.New().CronJobConfig.Prefix),
		CronJobConfigMixin{},
		mixin.NewDatetime(),
	}
}

// Indexes of the CronJobConfig.
func (CronJobConfig) Indexes() []ent.Index {
	return []ent.Index{
		// Unique job name
		index.Fields("job_name").Unique(),

		// Index for enabled jobs
		index.Fields("enabled"),

		// Index for job type lookups
		index.Fields("job_type"),
	}
}
