package controller

import (
	"context"
	"fmt"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/adapter/repository/cronjobconfigrepository"
	"sheng-go-backend/pkg/infrastructure/scheduler"
)

type CronJob interface {
	List(ctx context.Context) ([]*ent.CronJobConfig, error)
	GetByName(ctx context.Context, jobName string) (*ent.CronJobConfig, error)
	Update(ctx context.Context, jobName string, input ent.UpdateCronJobConfigInput) (*ent.CronJobConfig, error)
	Toggle(ctx context.Context, jobName string, enabled bool) (*ent.CronJobConfig, error)
}

type cronJobController struct {
	repo      *cronjobconfigrepository.CronJobConfigRepository
	scheduler *scheduler.Scheduler
}

func NewCronJobController(
	repo *cronjobconfigrepository.CronJobConfigRepository,
	scheduler *scheduler.Scheduler,
) CronJob {
	return &cronJobController{
		repo:      repo,
		scheduler: scheduler,
	}
}

func (c *cronJobController) List(ctx context.Context) ([]*ent.CronJobConfig, error) {
	return c.repo.List(ctx)
}

func (c *cronJobController) GetByName(ctx context.Context, jobName string) (*ent.CronJobConfig, error) {
	return c.repo.GetByName(ctx, jobName)
}

func (c *cronJobController) Update(ctx context.Context, jobName string, input ent.UpdateCronJobConfigInput) (*ent.CronJobConfig, error) {
	// Get current job to get its ID
	job, err := c.repo.GetByName(ctx, jobName)
	if err != nil {
		return nil, fmt.Errorf("failed to find job: %w", err)
	}

	// Build updates map
	updates := make(map[string]interface{})
	if input.Schedule != nil {
		updates["schedule"] = *input.Schedule
	}
	if input.Enabled != nil {
		updates["enabled"] = *input.Enabled
	}
	if input.BatchSize != nil {
		updates["batch_size"] = *input.BatchSize
	}
	if input.AdminEmail != nil {
		updates["admin_email"] = *input.AdminEmail
	}
	if input.RespectQuota != nil {
		updates["respect_quota"] = *input.RespectQuota
	}

	// Update the job
	updatedJob, err := c.repo.Update(ctx, string(job.ID), updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update job: %w", err)
	}

	// Reload schedule if the job is enabled
	if updatedJob.Enabled {
		if err := c.scheduler.ReloadSchedule(ctx, jobName); err != nil {
			return nil, fmt.Errorf("failed to reload schedule: %w", err)
		}
	}

	return updatedJob, nil
}

func (c *cronJobController) Toggle(ctx context.Context, jobName string, enabled bool) (*ent.CronJobConfig, error) {
	// Get current job to get its ID
	job, err := c.repo.GetByName(ctx, jobName)
	if err != nil {
		return nil, fmt.Errorf("failed to find job: %w", err)
	}

	// Toggle the job
	updatedJob, err := c.repo.Toggle(ctx, string(job.ID), enabled)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle job: %w", err)
	}

	// Reload schedule if enabled, or stop if disabled
	if enabled {
		if err := c.scheduler.ReloadSchedule(ctx, jobName); err != nil {
			return nil, fmt.Errorf("failed to reload schedule: %w", err)
		}
	} else {
		// When disabling, also reload (which will remove the job from cron)
		if err := c.scheduler.ReloadSchedule(ctx, jobName); err != nil {
			// If reload fails when disabling, it might be because job doesn't exist
			// Log but don't fail the operation
			fmt.Printf("Warning: Failed to reload schedule when disabling job: %v\n", err)
		}
	}

	return updatedJob, nil
}
