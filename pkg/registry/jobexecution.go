package registry

import (
	"sheng-go-backend/pkg/adapter/controller"
)

func (r *registry) NewJobExecutionController() controller.JobExecution {
	return controller.NewJobExecutionController(r.jobHistoryRepo, r.profileFetcher)
}
