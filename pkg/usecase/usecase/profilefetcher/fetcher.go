package profilefetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/jobexecutionhistory"
	"sheng-go-backend/ent/profileentry"
	"sheng-go-backend/pkg/adapter/repository/cronjobconfigrepository"
	"sheng-go-backend/pkg/adapter/repository/jobexecutionhistoryrepository"
	"sheng-go-backend/pkg/adapter/repository/profileentryrepository"
	"sheng-go-backend/pkg/adapter/repository/profilerepository"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/infrastructure/email"
	"sheng-go-backend/pkg/infrastructure/external/rapidapi"
	"sheng-go-backend/pkg/infrastructure/storage"
	"sheng-go-backend/pkg/usecase/usecase/apiquota"
	"strings"
	"time"
)

// ProfileFetcher handles the profile fetching workflow
type ProfileFetcher struct {
	profileEntryRepo profileentryrepository.ProfileEntryRepository
	profileRepo      profilerepository.ProfileRepository
	cronConfigRepo   *cronjobconfigrepository.CronJobConfigRepository
	jobHistoryRepo   *jobexecutionhistoryrepository.JobExecutionHistoryRepository
	linkedinClient   *rapidapi.LinkedInClient
	s3Service        *storage.S3Service
	emailService     *email.EmailService
	quotaManager     *apiquota.QuotaManager
}

// NewProfileFetcher creates a new ProfileFetcher
func NewProfileFetcher(
	profileEntryRepo profileentryrepository.ProfileEntryRepository,
	profileRepo profilerepository.ProfileRepository,
	cronConfigRepo *cronjobconfigrepository.CronJobConfigRepository,
	jobHistoryRepo *jobexecutionhistoryrepository.JobExecutionHistoryRepository,
	linkedinClient *rapidapi.LinkedInClient,
	s3Service *storage.S3Service,
	emailService *email.EmailService,
	quotaManager *apiquota.QuotaManager,
) *ProfileFetcher {
	return &ProfileFetcher{
		profileEntryRepo: profileEntryRepo,
		profileRepo:      profileRepo,
		cronConfigRepo:   cronConfigRepo,
		jobHistoryRepo:   jobHistoryRepo,
		linkedinClient:   linkedinClient,
		s3Service:        s3Service,
		emailService:     emailService,
		quotaManager:     quotaManager,
	}
}

// ExecuteFetchJob executes the profile fetching job
func (pf *ProfileFetcher) ExecuteFetchJob(ctx context.Context) (*ent.JobExecutionHistory, error) {
	startTime := time.Now()

	// Load cron job config
	jobConfig, err := pf.cronConfigRepo.GetByName(ctx, "profile_fetcher")
	if err != nil {
		return nil, fmt.Errorf("failed to get job config: %w", err)
	}

	// Check quota
	allowedBatchSize, err := pf.quotaManager.CheckAndReserveQuota(ctx, jobConfig.BatchSize)
	if err != nil {
		if jobConfig.RespectQuota {
			// Create history record for quota exceeded
			history := &ent.JobExecutionHistory{
				JobName:         "profile_fetcher",
				Status:          jobexecutionhistory.StatusQuotaExceeded,
				StartedAt:       startTime,
				CompletedAt:     ptr(time.Now()),
				TotalProcessed:  0,
				SuccessfulCount: 0,
				FailedCount:     0,
				APICallsMade:    0,
				DurationSeconds: int(time.Since(startTime).Seconds()),
			}
			errMsg := err.Error()
			history.ErrorSummary = &errMsg

			savedHistory, _ := pf.jobHistoryRepo.Create(ctx, history)
			return savedHistory, err
		}
		// If not respecting quota, log warning but continue
		fmt.Printf("Warning: Quota check failed but respect_quota is false: %v\n", err)
		allowedBatchSize = jobConfig.BatchSize
	}

	// Get pending profile entries
	pendingEntries, err := pf.profileEntryRepo.GetPendingBatch(ctx, allowedBatchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending entries: %w", err)
	}

	// Initialize tracking
	successCount := 0
	failedCount := 0
	apiCallsMade := 0
	var errors []string

	// Process each entry
	for _, entry := range pendingEntries {
		// Update status to FETCHING
		_, _ = pf.profileEntryRepo.UpdateStatus(
			ctx,
			string(entry.ID),
			profileentry.StatusFetching,
			nil,
		)

		// Fetch profile from RapidAPI
		profile, rawData, err := pf.linkedinClient.FetchProfileByURN(ctx, entry.LinkedinUrn)
		apiCallsMade++

		if err != nil {
			// Handle error
			errMsg := err.Error()
			_, _ = pf.profileEntryRepo.UpdateStatus(
				ctx,
				string(entry.ID),
				profileentry.StatusFAILED,
				&errMsg,
			)
			errors = append(errors, fmt.Sprintf("URN %s: %s", entry.LinkedinUrn, err.Error()))
			failedCount++
			continue
		}

		// Increment quota
		if err := pf.quotaManager.IncrementCallCount(ctx, 1); err != nil {
			fmt.Printf("Warning: Failed to increment quota: %v\n", err)
		}

		// Generate S3 keys
		timestamp := time.Now().Unix()
		rawS3Key := fmt.Sprintf("profiles/%s-%d-raw.json", entry.LinkedinUrn, timestamp)
		cleanedS3Key := fmt.Sprintf("profiles/%s-%d-cleaned.json", entry.LinkedinUrn, timestamp)

		// Upload raw JSON to S3
		if err := pf.s3Service.UploadJSON(ctx, rawS3Key, rawData); err != nil {
			errMsg := fmt.Sprintf("S3 upload failed: %v", err)
			_, _ = pf.profileEntryRepo.UpdateStatus(
				ctx,
				string(entry.ID),
				profileentry.StatusFAILED,
				&errMsg,
			)
			errors = append(errors, fmt.Sprintf("URN %s: %s", entry.LinkedinUrn, errMsg))
			failedCount++
			continue
		}

		// Extract and clean data
		cleanedData := pf.extractProfileData(profile)
		cleanedJSON, _ := json.Marshal(cleanedData)

		// Upload cleaned JSON to S3
		if err := pf.s3Service.UploadJSON(ctx, cleanedS3Key, cleanedJSON); err != nil {
			errMsg := fmt.Sprintf("S3 upload failed: %v", err)
			_, _ = pf.profileEntryRepo.UpdateStatus(
				ctx,
				string(entry.ID),
				profileentry.StatusFAILED,
				&errMsg,
			)
			errors = append(errors, fmt.Sprintf("URN %s: %s", entry.LinkedinUrn, errMsg))
			failedCount++
			continue
		}

		// Upsert profile in database
		dbProfile := pf.convertToDBProfile(profile, rawS3Key, cleanedS3Key)
		if _, err := pf.profileRepo.Upsert(ctx, dbProfile); err != nil {
			errMsg := fmt.Sprintf("DB upsert failed: %v", err)
			_, _ = pf.profileEntryRepo.UpdateStatus(
				ctx,
				string(entry.ID),
				profileentry.StatusFAILED,
				&errMsg,
			)
			errors = append(errors, fmt.Sprintf("URN %s: %s", entry.LinkedinUrn, errMsg))
			failedCount++
			continue
		}

		// Update profile entry as completed
		if _, err := pf.profileEntryRepo.UpdateAfterFetch(ctx, string(entry.ID), rawS3Key, cleanedS3Key); err != nil {
			fmt.Printf("Warning: Failed to update profile entry: %v\n", err)
		}

		successCount++
	}

	// Get current quota status
	quotaStatus, _ := pf.quotaManager.GetCurrentQuotaStatus(ctx)
	quotaRemaining := 0
	if quotaStatus != nil {
		quotaRemaining = quotaStatus.QuotaLimit - quotaStatus.CallCount
	}

	// Create job execution history
	completedAt := time.Now()
	duration := int(completedAt.Sub(startTime).Seconds())

	status := jobexecutionhistory.StatusSuccess
	if failedCount > 0 && successCount == 0 {
		status = jobexecutionhistory.StatusFailed
	} else if failedCount > 0 {
		status = jobexecutionhistory.StatusPartial
	}

	history := &ent.JobExecutionHistory{
		JobName:         "profile_fetcher",
		Status:          status,
		StartedAt:       startTime,
		CompletedAt:     &completedAt,
		TotalProcessed:  len(pendingEntries),
		SuccessfulCount: successCount,
		FailedCount:     failedCount,
		APICallsMade:    apiCallsMade,
		QuotaRemaining:  quotaRemaining,
		DurationSeconds: duration,
	}

	if len(errors) > 0 {
		errorSummary := strings.Join(errors, "; ")
		history.ErrorSummary = &errorSummary
	}

	savedHistory, err := pf.jobHistoryRepo.Create(ctx, history)
	if err != nil {
		fmt.Printf("Warning: Failed to create job history: %v\n", err)
	}

	// Send email summary
	nextRunTime := time.Now().Add(24 * time.Hour) // Placeholder
	if err := pf.emailService.SendJobCompletionSummary(
		"Profile Fetcher",
		duration,
		len(pendingEntries),
		successCount,
		failedCount,
		apiCallsMade,
		quotaRemaining,
		errors,
		nextRunTime,
	); err != nil {
		fmt.Printf("Warning: Failed to send email summary: %v\n", err)
	}

	return savedHistory, nil
}

// extractProfileData extracts relevant fields from RapidAPI profile
func (pf *ProfileFetcher) extractProfileData(
	profile *rapidapi.LinkedInProfile,
) map[string]interface{} {
	return map[string]interface{}{
		"urn":           profile.URN,
		"username":      profile.Username,
		"firstName":     profile.FirstName,
		"lastName":      profile.LastName,
		"headline":      profile.Headline,
		"geo":           profile.Geo,
		"educations":    profile.Educations,
		"fullPositions": profile.FullPositions,
		"skills":        profile.Skills,
	}
}

// convertToDBProfile converts RapidAPI profile to database profile
func (pf *ProfileFetcher) convertToDBProfile(
	profile *rapidapi.LinkedInProfile,
	rawS3Key, cleanedS3Key string,
) *ent.Profile {
	p := &ent.Profile{
		Urn:              profile.URN,
		Username:         &profile.Username,
		FirstName:        &profile.FirstName,
		LastName:         &profile.LastName,
		Headline:         &profile.Headline,
		Educations:       profile.Educations,
		Positions:        profile.FullPositions,
		Skills:           profile.Skills,
		RawDataS3Key:     &rawS3Key,
		CleanedDataS3Key: &cleanedS3Key,
	}

	if profile.Geo != nil {
		p.Country = &profile.Geo.CountryName
		p.City = &profile.Geo.CityName
		p.GeoData = map[string]interface{}{
			"country_name": profile.Geo.CountryName,
			"city_name":    profile.Geo.CityName,
		}
	}

	return p
}

// ptr is a helper to get pointer to value
func ptr[T any](v T) *T {
	return &v
}

func (pf *ProfileFetcher) fetchSingleProfileEntry(
	ctx context.Context,
	entry *model.ProfileEntry,
) error {
	// Update status to FETCHING
	_, _ = pf.profileEntryRepo.UpdateStatus(ctx, string(entry.ID), profileentry.StatusFetching, nil)

	// Fetch profile from RapidAPI
	profile, rawData, err := pf.linkedinClient.FetchProfileByURN(ctx, entry.LinkedinUrn)
	if err != nil {
		// Handle error
		errMsg := err.Error()
		_, _ = pf.profileEntryRepo.UpdateStatus(
			ctx,
			string(entry.ID),
			profileentry.StatusFAILED,
			&errMsg,
		)

		return err
	}

	// Increment quota
	if err := pf.quotaManager.IncrementCallCount(ctx, 1); err != nil {
		fmt.Printf("Warning: Failed to increment quota: %v\n", err)
	}

	// Generate S3 keys
	timestamp := time.Now().Unix()
	rawS3Key := fmt.Sprintf("profiles/%s-%d-raw.json", entry.LinkedinUrn, timestamp)
	cleanedS3Key := fmt.Sprintf("profiles/%s-%d-cleaned.json", entry.LinkedinUrn, timestamp)

	// Upload raw JSON to S3
	if err := pf.s3Service.UploadJSON(ctx, rawS3Key, rawData); err != nil {
		errMsg := fmt.Sprintf("S3 upload failed: %v", err)
		_, _ = pf.profileEntryRepo.UpdateStatus(
			ctx,
			string(entry.ID),
			profileentry.StatusFAILED,
			&errMsg,
		)
		return err
	}

	// Extract and clean data
	cleanedData := pf.extractProfileData(profile)
	cleanedJSON, _ := json.Marshal(cleanedData)

	// Upload cleaned JSON to S3
	if err := pf.s3Service.UploadJSON(ctx, cleanedS3Key, cleanedJSON); err != nil {
		errMsg := fmt.Sprintf("S3 upload failed: %v", err)
		_, _ = pf.profileEntryRepo.UpdateStatus(
			ctx,
			string(entry.ID),
			profileentry.StatusFAILED,
			&errMsg,
		)
		return err
	}

	// Upsert profile in database
	dbProfile := pf.convertToDBProfile(profile, rawS3Key, cleanedS3Key)
	if _, err := pf.profileRepo.Upsert(ctx, dbProfile); err != nil {
		errMsg := fmt.Sprintf("DB upsert failed: %v", err)
		_, _ = pf.profileEntryRepo.UpdateStatus(
			ctx,
			string(entry.ID),
			profileentry.StatusFAILED,
			&errMsg,
		)
		return err
	}

	// Update profile entry as completed
	if _, err := pf.profileEntryRepo.UpdateAfterFetch(ctx, string(entry.ID), rawS3Key, cleanedS3Key); err != nil {
		fmt.Printf("Warning: Failed to update profile entry: %v\n", err)
	}

	return nil
}

func (pf *ProfileFetcher) FetchSinglEntry(ctx context.Context, entryId model.ID) error {
	profileEntry, err := pf.profileEntryRepo.GetById(ctx, entryId)
	if err != nil {
		return err
	}

	err = pf.fetchSingleProfileEntry(ctx, profileEntry)
	if err != nil {
		return err
	}

	return nil
}
