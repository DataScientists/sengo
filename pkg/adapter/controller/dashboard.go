package controller

import (
	"context"
	"fmt"
	"sheng-go-backend/ent/profileentry"
	"sheng-go-backend/pkg/adapter/repository/cronjobconfigrepository"
	"sheng-go-backend/pkg/adapter/repository/jobexecutionhistoryrepository"
	"sheng-go-backend/pkg/adapter/repository/profileentryrepository"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/usecase/usecase/apiquota"
)

type Dashboard interface {
	GetOverview(ctx context.Context) (*model.DashboardOverview, error)
}

type dashboardController struct {
	quotaManager     *apiquota.QuotaManager
	profileEntryRepo profileentryrepository.ProfileEntryRepository
	cronJobRepo      *cronjobconfigrepository.CronJobConfigRepository
	jobHistoryRepo   *jobexecutionhistoryrepository.JobExecutionHistoryRepository
}

func NewDashboardController(
	quotaManager *apiquota.QuotaManager,
	profileEntryRepo profileentryrepository.ProfileEntryRepository,
	cronJobRepo *cronjobconfigrepository.CronJobConfigRepository,
	jobHistoryRepo *jobexecutionhistoryrepository.JobExecutionHistoryRepository,
) Dashboard {
	return &dashboardController{
		quotaManager:     quotaManager,
		profileEntryRepo: profileEntryRepo,
		cronJobRepo:      cronJobRepo,
		jobHistoryRepo:   jobHistoryRepo,
	}
}

func (c *dashboardController) GetOverview(ctx context.Context) (*model.DashboardOverview, error) {
	// Get quota status
	quotaStatus, err := c.quotaManager.GetCurrentQuotaStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota status: %w", err)
	}

	// Get pending profiles count
	pendingCount, err := c.profileEntryRepo.CountByStatus(ctx, profileentry.StatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending count: %w", err)
	}

	// Get recent job executions
	recentJobs, err := c.jobHistoryRepo.ListRecent(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent jobs: %w", err)
	}

	// Get cron jobs status
	cronJobs, err := c.cronJobRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cron jobs: %w", err)
	}

	// Get profile entry stats
	stats, err := c.profileEntryRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile entry stats: %w", err)
	}

	return &model.DashboardOverview{
		QuotaStatus:          quotaStatus,
		PendingProfilesCount: pendingCount,
		RecentJobExecutions:  recentJobs,
		CronJobsStatus:       cronJobs,
		ProfileEntryStats:    stats,
	}, nil
}
