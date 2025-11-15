package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sheng-go-backend/config"
	"sheng-go-backend/ent"
	_ "sheng-go-backend/ent/runtime"
	"sheng-go-backend/pkg/adapter/controller"
	"sheng-go-backend/pkg/adapter/repository/apiquotatrackerrepository"
	"sheng-go-backend/pkg/adapter/repository/cronjobconfigrepository"
	"sheng-go-backend/pkg/adapter/repository/jobexecutionhistoryrepository"
	"sheng-go-backend/pkg/adapter/repository/profileentryrepository"
	"sheng-go-backend/pkg/adapter/repository/profilerepository"
	"sheng-go-backend/pkg/infrastructure/datastore"
	"sheng-go-backend/pkg/infrastructure/email"
	"sheng-go-backend/pkg/infrastructure/external/rapidapi"
	"sheng-go-backend/pkg/infrastructure/graphql"
	"sheng-go-backend/pkg/infrastructure/router"
	"sheng-go-backend/pkg/infrastructure/scheduler"
	"sheng-go-backend/pkg/infrastructure/storage"
	"sheng-go-backend/pkg/registry"
	"sheng-go-backend/pkg/usecase/usecase/apiquota"
	"sheng-go-backend/pkg/usecase/usecase/profilefetcher"
	"syscall"
)

func main() {
	config.ReadConfig(config.ReadConfigOption{})

	client := newDBClient()

	// Initialize infrastructure services
	s3Service, err := storage.NewS3Service()
	if err != nil {
		log.Printf("Warning: Failed to initialize S3 service: %v", err)
	}

	emailService := email.NewEmailService()
	linkedinClient := rapidapi.NewLinkedInClient()

	// Initialize repositories
	profileEntryRepo := profileentryrepository.NewProfileEntryRepository(client)
	profileRepo := profilerepository.NewProfileRepo(client)
	quotaTrackerRepo := apiquotatrackerrepository.NewAPIQuotaTrackerRepository(client)
	cronConfigRepo := cronjobconfigrepository.NewCronJobConfigRepository(client)
	jobHistoryRepo := jobexecutionhistoryrepository.NewJobExecutionHistoryRepository(client)

	// Initialize usecases
	quotaManager := apiquota.NewQuotaManager(quotaTrackerRepo, emailService)
	profileFetcherUsecase := profilefetcher.NewProfileFetcher(
		profileEntryRepo,
		profileRepo,
		cronConfigRepo,
		jobHistoryRepo,
		linkedinClient,
		s3Service,
		emailService,
		quotaManager,
	)

	// Initialize and start cron scheduler
	cronScheduler := scheduler.NewScheduler(
		client,
		profileFetcherUsecase,
		quotaManager,
		cronConfigRepo,
	)

	// Initialize controller with all dependencies
	ctrl := newController(client, quotaManager, cronScheduler, profileFetcherUsecase, profileEntryRepo, cronConfigRepo, jobHistoryRepo)

	ctx := context.Background()
	if err := cronScheduler.Start(ctx); err != nil {
		log.Fatalf("Failed to start cron scheduler: %v", err)
	}
	defer cronScheduler.Stop()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		cronScheduler.Stop()
		os.Exit(0)
	}()

	// Start GraphQL server
	srv := graphql.NewServer(client, ctrl)

	e := router.New(srv, router.Options{
		Auth: false,
	})

	log.Printf("Starting server on :%s", config.C.Server.Address)
	e.Logger.Fatal(e.Start(":" + config.C.Server.Address))
}

func newDBClient() *ent.Client {
	client, err := datastore.NewClient()
	if err != nil {
		log.Fatalf("Failed to open db connection: %v", err)
	}
	return client
}

func newController(
	client *ent.Client,
	quotaManager *apiquota.QuotaManager,
	cronScheduler *scheduler.Scheduler,
	profileFetcher *profilefetcher.ProfileFetcher,
	profileEntryRepo profileentryrepository.ProfileEntryRepository,
	cronConfigRepo *cronjobconfigrepository.CronJobConfigRepository,
	jobHistoryRepo *jobexecutionhistoryrepository.JobExecutionHistoryRepository,
) controller.Controller {
	r := registry.NewWithOptions(client, registry.RegistryOptions{
		QuotaManager:     quotaManager,
		Scheduler:        cronScheduler,
		ProfileFetcher:   profileFetcher,
		ProfileEntryRepo: profileEntryRepo,
		CronConfigRepo:   cronConfigRepo,
		JobHistoryRepo:   jobHistoryRepo,
	})
	return r.NewController()
}
