package registry

import (
	"sheng-go-backend/pkg/adapter/controller"
)

func (r *registry) NewDashboardController() controller.Dashboard {
	return controller.NewDashboardController(
		r.quotaManager,
		r.profileEntryRepo,
		r.cronConfigRepo,
		r.jobHistoryRepo,
	)
}
