package controller

import (
	"context"
	"fmt"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/usecase/usecase/jobexecutionhistory"
	"sheng-go-backend/pkg/usecase/usecase/profilefetcher"
)

type JobExecution interface {
	List(ctx context.Context,
		after *model.Cursor,
		first *int,
		before *model.Cursor,
		last *int,
		where *ent.JobExecutionHistoryWhereInput,
	) (*ent.JobExecutionHistoryConnection, error)
	Get(ctx context.Context, id model.ID) (*ent.JobExecutionHistory, error)
	GetLatest(ctx context.Context, jobName string) (*ent.JobExecutionHistory, error)
	GetStats(ctx context.Context, jobName string, days int) (*model.JobStats, error)
	TriggerProfileFetch(ctx context.Context) (*ent.JobExecutionHistory, error)
}

type jobExecutionController struct {
	usecase        jobexecutionhistory.UseCase
	profileFetcher *profilefetcher.ProfileFetcher
}

func NewJobExecutionController(
	usecase jobexecutionhistory.UseCase,
	profileFetcher *profilefetcher.ProfileFetcher,
) JobExecution {
	return &jobExecutionController{
		usecase:        usecase,
		profileFetcher: profileFetcher,
	}
}

func (c *jobExecutionController) List(
	ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int,
	where *ent.JobExecutionHistoryWhereInput,
) (*ent.JobExecutionHistoryConnection, error) {
	return c.usecase.List(ctx, after, first, before, last, where)
}

func (c *jobExecutionController) Get(
	ctx context.Context,
	id model.ID,
) (*ent.JobExecutionHistory, error) {
	h, err := c.usecase.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get job execution history: %w", err)
	}
	return h, nil
}

func (c *jobExecutionController) GetLatest(
	ctx context.Context,
	jobName string,
) (*ent.JobExecutionHistory, error) {
	history, err := c.usecase.GetLatest(ctx, jobName)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest execution: %w", err)
	}
	return history, nil
}

func (c *jobExecutionController) GetStats(
	ctx context.Context,
	jobName string,
	days int,
) (*model.JobStats, error) {
	statsMap, err := c.usecase.GetStats(ctx, jobName, days)
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

func (c *jobExecutionController) TriggerProfileFetch(
	ctx context.Context,
) (*ent.JobExecutionHistory, error) {
	// Execute the profile fetch job
	history, err := c.profileFetcher.ExecuteFetchJob(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger profile fetch: %w", err)
	}

	return history, nil
}
