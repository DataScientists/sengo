//go:generate mockgen -source=jobexecutionhistory.go -destination=./mocks/jobexecutionhistory_repository_mock.go -package=mocks
package repository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/entity/model"
)

type JobExecutionHistory interface {
	List(ctx context.Context,
		after *model.Cursor,
		first *int,
		before *model.Cursor,
		last *int,
		where *ent.JobExecutionHistoryWhereInput,
	) (*ent.JobExecutionHistoryConnection, error)
	Get(ctx context.Context, id model.ID) (*ent.JobExecutionHistory, error)
	GetLatestByJobName(ctx context.Context, jobName string) (*ent.JobExecutionHistory, error)
	GetStats(ctx context.Context, jobName string, days int) (map[string]interface{}, error)
}
