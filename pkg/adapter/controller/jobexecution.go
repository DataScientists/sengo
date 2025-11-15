package controller

import (
	"context"
	"fmt"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/adapter/repository/jobexecutionhistoryrepository"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/usecase/usecase/profilefetcher"
)

type JobExecution interface {
	List(ctx context.Context, jobName *string, limit int, where *ent.JobExecutionHistoryWhereInput) ([]*ent.JobExecutionHistory, error)
	GetLatest(ctx context.Context, jobName string) (*ent.JobExecutionHistory, error)
	GetStats(ctx context.Context, jobName string, days int) (*model.JobStats, error)
	TriggerProfileFetch(ctx context.Context) (*ent.JobExecutionHistory, error)
}

type jobExecutionController struct {
	repo           *jobexecutionhistoryrepository.JobExecutionHistoryRepository
	profileFetcher *profilefetcher.ProfileFetcher
}

func NewJobExecutionController(
	repo *jobexecutionhistoryrepository.JobExecutionHistoryRepository,
	profileFetcher *profilefetcher.ProfileFetcher,
) JobExecution {
	return &jobExecutionController{
		repo:           repo,
		profileFetcher: profileFetcher,
	}
}

func (c *jobExecutionController) List(ctx context.Context, jobName *string, limit int, where *ent.JobExecutionHistoryWhereInput) ([]*ent.JobExecutionHistory, error) {
	// If jobName is specified, filter by job name
	if jobName != nil {
		return c.repo.ListByJobName(ctx, *jobName, limit)
	}

	// Otherwise return recent executions across all jobs
	return c.repo.ListRecent(ctx, limit)
}

func (c *jobExecutionController) GetLatest(ctx context.Context, jobName string) (*ent.JobExecutionHistory, error) {
	history, err := c.repo.GetLatestByJobName(ctx, jobName)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest execution: %w", err)
	}
	return history, nil
}

func (c *jobExecutionController) GetStats(ctx context.Context, jobName string, days int) (*model.JobStats, error) {
	statsMap, err := c.repo.GetStats(ctx, jobName, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get job stats: %w", err)
	}

	// Convert map to JobStats struct
	stats := &model.JobStats{
		TotalExecutions:   statsMap["total_executions"].(int),
		SuccessRate:       statsMap["success_rate"].(float64),
		AverageDuration:   statsMap["average_duration"].(int),
		TotalProfiles:     statsMap["total_profiles"].(int),
		TotalAPICallsMade: statsMap["total_api_calls_made"].(int),
	}

	return stats, nil
}

func (c *jobExecutionController) TriggerProfileFetch(ctx context.Context) (*ent.JobExecutionHistory, error) {
	// Execute the profile fetch job
	history, err := c.profileFetcher.ExecuteFetchJob(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger profile fetch: %w", err)
	}

	return history, nil
}
