package registry

import (
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/adapter/controller"
	"sheng-go-backend/pkg/adapter/repository/cronjobconfigrepository"
	"sheng-go-backend/pkg/adapter/repository/jobexecutionhistoryrepository"
	"sheng-go-backend/pkg/adapter/repository/profileentryrepository"
	"sheng-go-backend/pkg/infrastructure/scheduler"
	"sheng-go-backend/pkg/usecase/usecase/apiquota"
	"sheng-go-backend/pkg/usecase/usecase/profilefetcher"
)

type registry struct {
	client               *ent.Client
	quotaManager         *apiquota.QuotaManager
	scheduler            *scheduler.Scheduler
	profileFetcher       *profilefetcher.ProfileFetcher
	profileEntryRepo     profileentryrepository.ProfileEntryRepository
	cronConfigRepo       *cronjobconfigrepository.CronJobConfigRepository
	jobHistoryRepo       *jobexecutionhistoryrepository.JobExecutionHistoryRepository
}

// Registry is an interface of registry
type Registry interface {
	NewController() controller.Controller
}

// RegistryOptions contains optional dependencies for registry
type RegistryOptions struct {
	QuotaManager     *apiquota.QuotaManager
	Scheduler        *scheduler.Scheduler
	ProfileFetcher   *profilefetcher.ProfileFetcher
	ProfileEntryRepo profileentryrepository.ProfileEntryRepository
	CronConfigRepo   *cronjobconfigrepository.CronJobConfigRepository
	JobHistoryRepo   *jobexecutionhistoryrepository.JobExecutionHistoryRepository
}

// New registers entire controller with dependencies
func New(client *ent.Client) Registry {
	return &registry{client: client}
}

// NewWithOptions registers entire controller with additional dependencies
func NewWithOptions(client *ent.Client, opts RegistryOptions) Registry {
	return &registry{
		client:           client,
		quotaManager:     opts.QuotaManager,
		scheduler:        opts.Scheduler,
		profileFetcher:   opts.ProfileFetcher,
		profileEntryRepo: opts.ProfileEntryRepo,
		cronConfigRepo:   opts.CronConfigRepo,
		jobHistoryRepo:   opts.JobHistoryRepo,
	}
}

// NewController generates controllers
func (r *registry) NewController() controller.Controller {
	return controller.Controller{
		User:         r.NewUserController(),
		Todo:         r.NewTodoController(),
		Profile:      r.NewProfileController(),
		Auth:         r.NewAuthController(),
		ProfileEntry: r.NewProfileEntryController(),
		APIQuota:     r.NewAPIQuotaController(),
		CronJob:      r.NewCronJobController(),
		JobExecution: r.NewJobExecutionController(),
		Dashboard:    r.NewDashboardController(),
	}
}
