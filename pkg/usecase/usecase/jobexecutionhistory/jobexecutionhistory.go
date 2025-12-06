package jobexecutionhistory

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/usecase/repository"
)

type useCase struct {
	repo repository.JobExecutionHistory
}

type UseCase interface {
	List(ctx context.Context,
		after *model.Cursor,
		first *int,
		before *model.Cursor,
		last *int,
		where *ent.JobExecutionHistoryWhereInput,
	) (*ent.JobExecutionHistoryConnection, error)
	Get(ctx context.Context, id model.ID) (*ent.JobExecutionHistory, error)
	GetLatest(ctx context.Context, jobName string) (*ent.JobExecutionHistory, error)
	GetStats(ctx context.Context, jobName string, days int) (map[string]interface{}, error)
}

func New(repo repository.JobExecutionHistory) UseCase {
	return &useCase{repo: repo}
}

func (u *useCase) List(
	ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int,
	where *ent.JobExecutionHistoryWhereInput,
) (*ent.JobExecutionHistoryConnection, error) {
	return u.repo.List(ctx, after, first, before, last, where)
}

func (u *useCase) Get(ctx context.Context, id model.ID) (*ent.JobExecutionHistory, error) {
	return u.repo.Get(ctx, id)
}

func (u *useCase) GetLatest(ctx context.Context, jobName string) (*ent.JobExecutionHistory, error) {
	return u.repo.GetLatestByJobName(ctx, jobName)
}

func (u *useCase) GetStats(ctx context.Context, jobName string, days int) (map[string]interface{}, error) {
	return u.repo.GetStats(ctx, jobName, days)
}
