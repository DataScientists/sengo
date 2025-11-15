package jobexecutionhistoryrepository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/jobexecutionhistory"
)

type JobExecutionHistoryRepository struct {
	client *ent.Client
}

func NewJobExecutionHistoryRepository(client *ent.Client) *JobExecutionHistoryRepository {
	return &JobExecutionHistoryRepository{client: client}
}

// Create creates a new job execution history record
func (r *JobExecutionHistoryRepository) Create(ctx context.Context, input *ent.JobExecutionHistory) (*ent.JobExecutionHistory, error) {
	builder := r.client.JobExecutionHistory.
		Create().
		SetJobName(input.JobName).
		SetStatus(input.Status).
		SetStartedAt(input.StartedAt).
		SetTotalProcessed(input.TotalProcessed).
		SetSuccessfulCount(input.SuccessfulCount).
		SetFailedCount(input.FailedCount).
		SetAPICallsMade(input.APICallsMade).
		SetQuotaRemaining(input.QuotaRemaining).
		SetDurationSeconds(input.DurationSeconds)

	if input.CompletedAt != nil {
		builder = builder.SetCompletedAt(*input.CompletedAt)
	}
	if input.ErrorSummary != nil {
		builder = builder.SetErrorSummary(*input.ErrorSummary)
	}

	return builder.Save(ctx)
}

// GetLatestByJobName retrieves the most recent execution for a job
func (r *JobExecutionHistoryRepository) GetLatestByJobName(ctx context.Context, jobName string) (*ent.JobExecutionHistory, error) {
	return r.client.JobExecutionHistory.
		Query().
		Where(jobexecutionhistory.JobName(jobName)).
		Order(ent.Desc(jobexecutionhistory.FieldStartedAt)).
		First(ctx)
}

// ListByJobName retrieves execution history for a specific job
func (r *JobExecutionHistoryRepository) ListByJobName(ctx context.Context, jobName string, limit int) ([]*ent.JobExecutionHistory, error) {
	query := r.client.JobExecutionHistory.
		Query().
		Where(jobexecutionhistory.JobName(jobName)).
		Order(ent.Desc(jobexecutionhistory.FieldStartedAt))

	if limit > 0 {
		query = query.Limit(limit)
	}

	return query.All(ctx)
}

// ListRecent retrieves recent execution history across all jobs
func (r *JobExecutionHistoryRepository) ListRecent(ctx context.Context, limit int) ([]*ent.JobExecutionHistory, error) {
	return r.client.JobExecutionHistory.
		Query().
		Order(ent.Desc(jobexecutionhistory.FieldStartedAt)).
		Limit(limit).
		All(ctx)
}

// ListByStatus retrieves executions by status
func (r *JobExecutionHistoryRepository) ListByStatus(ctx context.Context, status jobexecutionhistory.Status, limit int) ([]*ent.JobExecutionHistory, error) {
	query := r.client.JobExecutionHistory.
		Query().
		Where(jobexecutionhistory.StatusEQ(status)).
		Order(ent.Desc(jobexecutionhistory.FieldStartedAt))

	if limit > 0 {
		query = query.Limit(limit)
	}

	return query.All(ctx)
}

// GetStats retrieves statistics for a job
func (r *JobExecutionHistoryRepository) GetStats(ctx context.Context, jobName string, days int) (map[string]interface{}, error) {
	// This would typically use aggregate functions
	// For now, returning basic structure
	executions, err := r.ListByJobName(ctx, jobName, 100)
	if err != nil {
		return nil, err
	}

	totalExecs := len(executions)
	successCount := 0
	totalDuration := 0
	totalProfiles := 0
	totalAPICalls := 0

	for _, exec := range executions {
		if exec.Status == jobexecutionhistory.StatusSuccess {
			successCount++
		}
		totalDuration += exec.DurationSeconds
		totalProfiles += exec.SuccessfulCount
		totalAPICalls += exec.APICallsMade
	}

	successRate := 0.0
	if totalExecs > 0 {
		successRate = float64(successCount) / float64(totalExecs) * 100
	}

	avgDuration := 0
	if totalExecs > 0 {
		avgDuration = totalDuration / totalExecs
	}

	return map[string]interface{}{
		"total_executions":     totalExecs,
		"success_rate":         successRate,
		"average_duration":     avgDuration,
		"total_profiles":       totalProfiles,
		"total_api_calls_made": totalAPICalls,
	}, nil
}
