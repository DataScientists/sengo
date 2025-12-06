package registry

import (
	"sheng-go-backend/pkg/adapter/controller"
	jobexecutionhistoryrepository "sheng-go-backend/pkg/adapter/repository/jobexecutionhistoryrepository"
	jobexecutionhistoryusecase "sheng-go-backend/pkg/usecase/usecase/jobexecutionhistory"
)

func (r *registry) NewJobExecutionController() controller.JobExecution {
	repo := r.jobHistoryRepo
	if repo == nil {
		repo = jobexecutionhistoryrepository.NewJobExecutionHistoryRepository(r.client)
	}
	usecase := jobexecutionhistoryusecase.New(repo)

	return controller.NewJobExecutionController(usecase, r.profileFetcher)
}
