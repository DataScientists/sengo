package scheduler

import (
	"context"
	"fmt"
	"log"
	"sheng-go-backend/config"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/cronjobconfig"
	"sheng-go-backend/pkg/adapter/repository/cronjobconfigrepository"
	"sheng-go-backend/pkg/usecase/usecase/apiquota"
	"sheng-go-backend/pkg/usecase/usecase/profilefetcher"

	"github.com/robfig/cron/v3"
)

// Scheduler manages cron jobs
type Scheduler struct {
	cron           *cron.Cron
	client         *ent.Client
	profileFetcher *profilefetcher.ProfileFetcher
	quotaManager   *apiquota.QuotaManager
	cronRepo       *cronjobconfigrepository.CronJobConfigRepository
	entryIDs       map[string]cron.EntryID // Map job names to cron entry IDs
}

// NewScheduler creates a new scheduler
func NewScheduler(
	client *ent.Client,
	profileFetcher *profilefetcher.ProfileFetcher,
	quotaManager *apiquota.QuotaManager,
	cronRepo *cronjobconfigrepository.CronJobConfigRepository,
) *Scheduler {
	return &Scheduler{
		cron:           cron.New(),
		client:         client,
		profileFetcher: profileFetcher,
		quotaManager:   quotaManager,
		cronRepo:       cronRepo,
		entryIDs:       make(map[string]cron.EntryID),
	}
}

// Start initializes and starts all cron jobs
func (s *Scheduler) Start(ctx context.Context) error {
	log.Println("Starting cron scheduler...")

	// Initialize default cron job configs if they don't exist
	if err := s.initializeDefaultConfigs(ctx); err != nil {
		return fmt.Errorf("failed to initialize default configs: %w", err)
	}

	// Load enabled jobs from database
	jobs, err := s.cronRepo.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("failed to load cron jobs: %w", err)
	}

	// Register each job
	for _, job := range jobs {
		if err := s.registerJob(ctx, job); err != nil {
			log.Printf("Warning: Failed to register job %s: %v", job.JobName, err)
		}
	}

	// Start the cron scheduler
	s.cron.Start()
	log.Println("Cron scheduler started successfully")

	return nil
}

// Stop stops the cron scheduler
func (s *Scheduler) Stop() {
	log.Println("Stopping cron scheduler...")
	cronCtx := s.cron.Stop()
	<-cronCtx.Done()
	log.Println("Cron scheduler stopped")
}

// registerJob registers a single cron job
func (s *Scheduler) registerJob(ctx context.Context, job *ent.CronJobConfig) error {
	var entryID cron.EntryID
	var err error

	switch job.JobType {
	case cronjobconfig.JobTypeProfileFetcher:
		entryID, err = s.cron.AddFunc(job.Schedule, func() {
			s.runProfileFetcherJob(context.Background())
		})
	case cronjobconfig.JobTypeQuotaReset:
		entryID, err = s.cron.AddFunc(job.Schedule, func() {
			s.runQuotaResetJob(context.Background())
		})
	default:
		return fmt.Errorf("unknown job type: %s", job.JobType)
	}

	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.entryIDs[job.JobName] = entryID
	log.Printf("Registered job: %s with schedule: %s", job.JobName, job.Schedule)

	return nil
}

// runProfileFetcherJob executes the profile fetcher job
func (s *Scheduler) runProfileFetcherJob(ctx context.Context) {
	log.Println("Running profile fetcher job...")

	// Update last run time
	if err := s.updateLastRun(ctx, "profile_fetcher"); err != nil {
		log.Printf("Warning: Failed to update last run time: %v", err)
	}

	// Execute the job
	history, err := s.profileFetcher.ExecuteFetchJob(ctx)
	if err != nil {
		log.Printf("Profile fetcher job failed: %v", err)
		return
	}

	log.Printf("Profile fetcher job completed: %d successful, %d failed, %d API calls",
		history.SuccessfulCount, history.FailedCount, history.APICallsMade)
}

// runQuotaResetJob executes the quota reset job
func (s *Scheduler) runQuotaResetJob(ctx context.Context) {
	log.Println("Running quota reset job...")

	// Update last run time
	if err := s.updateLastRun(ctx, "quota_reset"); err != nil {
		log.Printf("Warning: Failed to update last run time: %v", err)
	}

	// Execute the job
	if err := s.quotaManager.ResetMonthlyQuota(ctx); err != nil {
		log.Printf("Quota reset job failed: %v", err)
		return
	}

	log.Println("Quota reset job completed successfully")
}

// updateLastRun updates the last run timestamp for a job
func (s *Scheduler) updateLastRun(ctx context.Context, jobName string) error {
	job, err := s.cronRepo.GetByName(ctx, jobName)
	if err != nil {
		return err
	}

	_, err = s.cronRepo.UpdateLastRun(ctx, string(job.ID))
	return err
}

// initializeDefaultConfigs creates default cron job configurations
func (s *Scheduler) initializeDefaultConfigs(ctx context.Context) error {
	cfg := config.C

	// Profile Fetcher Job
	_, err := s.cronRepo.GetByName(ctx, "profile_fetcher")
	if err != nil && ent.IsNotFound(err) {
		log.Println("Creating default profile_fetcher job config...")
		_, err = s.cronRepo.Create(ctx, &ent.CronJobConfig{
			JobName:      "profile_fetcher",
			JobType:      cronjobconfig.JobTypeProfileFetcher,
			Schedule:     cfg.Cron.ProfileFetcherSchedule,
			Enabled:      true,
			BatchSize:    cfg.Cron.BatchSize,
			AdminEmail:   cfg.Email.AdminEmail,
			RespectQuota: true,
		})
		if err != nil {
			return fmt.Errorf("failed to create profile_fetcher config: %w", err)
		}
	}

	// Quota Reset Job
	_, err = s.cronRepo.GetByName(ctx, "quota_reset")
	if err != nil && ent.IsNotFound(err) {
		log.Println("Creating default quota_reset job config...")
		_, err = s.cronRepo.Create(ctx, &ent.CronJobConfig{
			JobName:      "quota_reset",
			JobType:      cronjobconfig.JobTypeQuotaReset,
			Schedule:     cfg.Cron.QuotaResetSchedule,
			Enabled:      true,
			BatchSize:    1,
			AdminEmail:   cfg.Email.AdminEmail,
			RespectQuota: false,
		})
		if err != nil {
			return fmt.Errorf("failed to create quota_reset config: %w", err)
		}
	}

	return nil
}

// ReloadSchedule reloads the schedule for a specific job (used when updating via dashboard)
func (s *Scheduler) ReloadSchedule(ctx context.Context, jobName string) error {
	// Remove existing entry
	if entryID, ok := s.entryIDs[jobName]; ok {
		s.cron.Remove(entryID)
		delete(s.entryIDs, jobName)
	}

	// Load updated config
	job, err := s.cronRepo.GetByName(ctx, jobName)
	if err != nil {
		return fmt.Errorf("failed to load job config: %w", err)
	}

	// Re-register job
	return s.registerJob(ctx, job)
}
