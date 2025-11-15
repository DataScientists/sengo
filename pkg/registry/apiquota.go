package registry

import (
	"sheng-go-backend/pkg/adapter/controller"
)

func (r *registry) NewAPIQuotaController() controller.APIQuota {
	return controller.NewAPIQuotaController(r.quotaManager)
}
