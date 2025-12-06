package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sheng-go-backend/config"
	"sheng-go-backend/pkg/adapter/repository/apiquotatrackerrepository"
	"sheng-go-backend/pkg/adapter/repository/cronjobconfigrepository"
	"sheng-go-backend/pkg/adapter/repository/jobexecutionhistoryrepository"
	"sheng-go-backend/pkg/adapter/repository/profileentryrepository"
	"sheng-go-backend/pkg/adapter/repository/profilerepository"
	"sheng-go-backend/pkg/infrastructure/datastore"
	"sheng-go-backend/pkg/infrastructure/email"
	"sheng-go-backend/pkg/infrastructure/external/rapidapi"
	"sheng-go-backend/pkg/infrastructure/storage"
	"sheng-go-backend/pkg/registry"
	"sheng-go-backend/pkg/usecase/usecase/apiquota"
	"sheng-go-backend/pkg/usecase/usecase/profilefetcher"
)

func main() {
	config.ReadConfig(config.ReadConfigOption{})

	client, err := datastore.NewClient()
	if err != nil {
		log.Fatalf("failed to open db connection: %v", err)
	}
	defer client.Close()

	// Infrastructure services
	s3Service, err := storage.NewS3Service()
	if err != nil {
		log.Fatalf("failed to initialize S3 service: %v", err)
	}
	emailService := email.NewEmailService()
	linkedinClient := rapidapi.NewLinkedInClient()

	// Repositories
	profileEntryRepo := profileentryrepository.NewProfileEntryRepository(client)
	profileRepo := profilerepository.NewProfileRepo(client)
	quotaTrackerRepo := apiquotatrackerrepository.NewAPIQuotaTrackerRepository(client)
	cronConfigRepo := cronjobconfigrepository.NewCronJobConfigRepository(client)
	jobHistoryRepo := jobexecutionhistoryrepository.NewJobExecutionHistoryRepository(client)

	// Usecases
	quotaManager := apiquota.NewQuotaManager(quotaTrackerRepo, emailService)
	profileFetcher := profilefetcher.NewProfileFetcher(
		profileEntryRepo,
		profileRepo,
		cronConfigRepo,
		jobHistoryRepo,
		linkedinClient,
		s3Service,
		emailService,
		quotaManager,
	)

	reg := registry.NewWithOptions(client, registry.RegistryOptions{
		ProfileFetcher:   profileFetcher,
		ProfileEntryRepo: profileEntryRepo,
		CronConfigRepo:   cronConfigRepo,
		JobHistoryRepo:   jobHistoryRepo,
		QuotaManager:     quotaManager,
	})
	ctrl := reg.NewController()

	ctx := context.Background()
	history, err := ctrl.JobExecution.TriggerProfileFetch(ctx)
	if err != nil {
		log.Fatalf("failed to trigger profile fetch job: %v", err)
	}

	fmt.Printf(
		"Profile fetch job triggered successfully. Job ID: %s, status: %s, processed: %d, success: %d, failed: %d, apiCalls: %d\n",
		history.ID,
		history.Status,
		history.TotalProcessed,
		history.SuccessfulCount,
		history.FailedCount,
		history.APICallsMade,
	)

	os.Exit(0)
}
