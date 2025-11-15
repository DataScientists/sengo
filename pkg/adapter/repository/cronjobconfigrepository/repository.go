package cronjobconfigrepository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/cronjobconfig"
	"sheng-go-backend/ent/schema/ulid"
	"time"
)

type CronJobConfigRepository struct {
	client *ent.Client
}

func NewCronJobConfigRepository(client *ent.Client) *CronJobConfigRepository {
	return &CronJobConfigRepository{client: client}
}

// GetByName retrieves a cron job config by job name
func (r *CronJobConfigRepository) GetByName(ctx context.Context, jobName string) (*ent.CronJobConfig, error) {
	return r.client.CronJobConfig.
		Query().
		Where(cronjobconfig.JobName(jobName)).
		Only(ctx)
}

// List retrieves all cron job configs
func (r *CronJobConfigRepository) List(ctx context.Context) ([]*ent.CronJobConfig, error) {
	return r.client.CronJobConfig.
		Query().
		All(ctx)
}

// ListEnabled retrieves all enabled cron job configs
func (r *CronJobConfigRepository) ListEnabled(ctx context.Context) ([]*ent.CronJobConfig, error) {
	return r.client.CronJobConfig.
		Query().
		Where(cronjobconfig.Enabled(true)).
		All(ctx)
}

// Create creates a new cron job config
func (r *CronJobConfigRepository) Create(ctx context.Context, input *ent.CronJobConfig) (*ent.CronJobConfig, error) {
	return r.client.CronJobConfig.
		Create().
		SetJobName(input.JobName).
		SetJobType(input.JobType).
		SetSchedule(input.Schedule).
		SetEnabled(input.Enabled).
		SetBatchSize(input.BatchSize).
		SetAdminEmail(input.AdminEmail).
		SetRespectQuota(input.RespectQuota).
		Save(ctx)
}

// Update updates a cron job config
func (r *CronJobConfigRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*ent.CronJobConfig, error) {
	updateQuery := r.client.CronJobConfig.UpdateOneID(ulid.ID(id))

	if schedule, ok := updates["schedule"].(string); ok {
		updateQuery = updateQuery.SetSchedule(schedule)
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		updateQuery = updateQuery.SetEnabled(enabled)
	}
	if batchSize, ok := updates["batch_size"].(int); ok {
		updateQuery = updateQuery.SetBatchSize(batchSize)
	}
	if adminEmail, ok := updates["admin_email"].(string); ok {
		updateQuery = updateQuery.SetAdminEmail(adminEmail)
	}
	if respectQuota, ok := updates["respect_quota"].(bool); ok {
		updateQuery = updateQuery.SetRespectQuota(respectQuota)
	}

	return updateQuery.Save(ctx)
}

// UpdateLastRun updates the last run timestamp
func (r *CronJobConfigRepository) UpdateLastRun(ctx context.Context, id string) (*ent.CronJobConfig, error) {
	return r.client.CronJobConfig.
		UpdateOneID(ulid.ID(id)).
		SetLastRunAt(time.Now()).
		Save(ctx)
}

// UpdateNextRun updates the next run timestamp
func (r *CronJobConfigRepository) UpdateNextRun(ctx context.Context, id string, nextRun time.Time) (*ent.CronJobConfig, error) {
	return r.client.CronJobConfig.
		UpdateOneID(ulid.ID(id)).
		SetNextRunAt(nextRun).
		Save(ctx)
}

// Toggle enables/disables a job
func (r *CronJobConfigRepository) Toggle(ctx context.Context, id string, enabled bool) (*ent.CronJobConfig, error) {
	return r.client.CronJobConfig.
		UpdateOneID(ulid.ID(id)).
		SetEnabled(enabled).
		Save(ctx)
}
