package registry

import (
	"sheng-go-backend/pkg/adapter/controller"
)

func (r *registry) NewCronJobController() controller.CronJob {
	return controller.NewCronJobController(r.cronConfigRepo, r.scheduler)
}
