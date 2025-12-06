package profilefetcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sheng-go-backend/config"
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

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	logger           *zap.SugaredLogger
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
		logger:           newProfileFetcherLogger(),
	}
}

// ExecuteFetchJob executes the profile fetching job
func (pf *ProfileFetcher) ExecuteFetchJob(ctx context.Context) (*ent.JobExecutionHistory, error) {
	startTime := time.Now()
	pf.logger.Info("profile fetcher job started")

	// Load cron job config
	jobConfig, err := pf.cronConfigRepo.GetByName(ctx, "profile_fetcher")
	if err != nil {
		return nil, fmt.Errorf("failed to get job config: %w", err)
	}

	// Check quota
	// Initialize tracking
	successCount := 0
	failedCount := 0
	apiCallsMade := 0
	totalProcessed := 0
	quotaLimited := false
	var errors []string

	for {
		allowedBatchSize, err := pf.quotaManager.CheckAndReserveQuota(ctx, jobConfig.BatchSize)
		if err != nil {
			if totalProcessed == 0 && jobConfig.RespectQuota {
				pf.logger.Warnw("quota exceeded before processing", "error", err)
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

			if jobConfig.RespectQuota {
				quotaLimited = true
				errors = append(errors, fmt.Sprintf("Stopped due to quota: %v", err))
				pf.logger.Warnw("stopping due to quota mid-run", "error", err)
				break
			}

			// If not respecting quota, log warning but continue
			pf.logger.Warnw(
				"quota check failed but respect_quota is false, continuing",
				"error",
				err,
			)
			allowedBatchSize = jobConfig.BatchSize
		}

		pf.logger.Infow("quota check passed", "allowed_batch_size", allowedBatchSize)

		// Get pending profile entries for this batch
		pendingEntries, err := pf.profileEntryRepo.GetPendingBatch(ctx, allowedBatchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to get pending entries: %w", err)
		}

		if len(pendingEntries) == 0 {
			pf.logger.Info("no pending profile entries, exiting")
			break
		}

		// Process each entry in the batch
		for _, entry := range pendingEntries {
			pf.logger.Infow("fetching profile", "urn", entry.LinkedinUrn)
			// Update status to FETCHING
			_, _ = pf.profileEntryRepo.UpdateStatus(
				ctx,
				string(entry.ID),
				profileentry.StatusFetching,
				nil,
			)

			// Fetch profile from RapidAPI
			profile, rawData, attempts, err := pf.fetchProfileWithRetry(ctx, entry.LinkedinUrn)
			apiCallsMade += attempts

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

			pf.logger.Infow(
				"fetched profile",
				"urn",
				entry.LinkedinUrn,
				"username",
				profile.Username,
			)

			// Increment quota
			if err := pf.quotaManager.IncrementCallCount(ctx, 1); err != nil {
				pf.logger.Warnw("failed to increment quota", "error", err)
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
				pf.logger.Errorw(
					"failed to upload raw json to s3",
					"urn",
					entry.LinkedinUrn,
					"error",
					err,
				)
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
				pf.logger.Errorw(
					"failed to upload cleaned json to s3",
					"urn",
					entry.LinkedinUrn,
					"error",
					err,
				)
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
				pf.logger.Errorw("failed to upsert profile", "urn", entry.LinkedinUrn, "error", err)
				continue
			}

			// Update profile entry as completed
			if _, err := pf.profileEntryRepo.UpdateAfterFetch(ctx, string(entry.ID), rawS3Key, cleanedS3Key); err != nil {
				pf.logger.Warnw(
					"failed to update profile entry after fetch",
					"urn",
					entry.LinkedinUrn,
					"error",
					err,
				)
			}

			successCount++
		}

		totalProcessed += len(pendingEntries)
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
	} else if failedCount > 0 || quotaLimited {
		status = jobexecutionhistory.StatusPartial
	}

	history := &ent.JobExecutionHistory{
		JobName:         "profile_fetcher",
		Status:          status,
		StartedAt:       startTime,
		CompletedAt:     &completedAt,
		TotalProcessed:  totalProcessed,
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

	if totalProcessed == 0 && !quotaLimited {
		// Nothing to do; skip persisting history to avoid noisy records.
		return history, nil
	}

	savedHistory, err := pf.jobHistoryRepo.Create(ctx, history)
	if err != nil {
		pf.logger.Warnw("failed to create job history", "error", err)
		// Return in-memory history even if persistence failed
		return history, nil
	}

	// Send email summary
	nextRunTime := time.Now().Add(24 * time.Hour) // Placeholder
	if err := pf.emailService.SendJobCompletionSummary(
		"Profile Fetcher",
		duration,
		totalProcessed,
		successCount,
		failedCount,
		apiCallsMade,
		quotaRemaining,
		errors,
		nextRunTime,
	); err != nil {
		pf.logger.Warnw("failed to send job completion email", "error", err)
	}

	return savedHistory, nil
}

func (pf *ProfileFetcher) fetchProfileWithRetry(
	ctx context.Context,
	urn string,
) (*rapidapi.LinkedInProfile, []byte, int, error) {
	cfg := config.C.RapidAPI

	maxRetries := cfg.RateLimitMaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	backoff := time.Duration(cfg.RateLimitBackoffMs) * time.Millisecond
	if backoff <= 0 {
		backoff = time.Second
	}

	maxBackoff := time.Duration(cfg.RateLimitBackoffMaxMs) * time.Millisecond
	if maxBackoff <= 0 {
		maxBackoff = 8 * time.Second
	}

	attempts := 0
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		attempts++

		profile, rawData, err := pf.linkedinClient.FetchProfileByURN(ctx, urn)
		if err == nil {
			return profile, rawData, attempts, nil
		}

		lastErr = err

		var rateErr *rapidapi.RateLimitError
		if errors.As(err, &rateErr) {
			if attempt == maxRetries {
				break
			}

			sleep := backoff
			if rateErr.RetryAfter > 0 && rateErr.RetryAfter > sleep {
				sleep = rateErr.RetryAfter
			}
			if maxBackoff > 0 && sleep > maxBackoff {
				sleep = maxBackoff
			}

			pf.logger.Infow(
				"rate limited, backing off",
				"urn",
				urn,
				"sleep",
				sleep,
				"attempt",
				attempt+1,
				"max_retries",
				maxRetries,
			)
			if err := sleepWithContext(ctx, sleep); err != nil {
				return nil, nil, attempts, err
			}

			backoff *= 2
			if maxBackoff > 0 && backoff > maxBackoff {
				backoff = maxBackoff
			}

			continue
		}

		// Non-rate-limit error, return immediately
		return nil, nil, attempts, err
	}

	return nil, nil, attempts, lastErr
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
		Title:            &profile.Headline,
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
	profile, rawData, _, err := pf.fetchProfileWithRetry(ctx, entry.LinkedinUrn)
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
		pf.logger.Warnw("failed to increment quota", "urn", entry.LinkedinUrn, "error", err)
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
		pf.logger.Errorw("failed to upload raw json to s3", "urn", entry.LinkedinUrn, "error", err)
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
		pf.logger.Errorw(
			"failed to upload cleaned json to s3",
			"urn",
			entry.LinkedinUrn,
			"error",
			err,
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
		pf.logger.Errorw("failed to upsert profile", "urn", entry.LinkedinUrn, "error", err)
		return err
	}

	// Update profile entry as completed
	if _, err := pf.profileEntryRepo.UpdateAfterFetch(ctx, string(entry.ID), rawS3Key, cleanedS3Key); err != nil {
		pf.logger.Warnw(
			"failed to update profile entry after fetch",
			"urn",
			entry.LinkedinUrn,
			"error",
			err,
		)
	}

	return nil
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func newProfileFetcherLogger() *zap.SugaredLogger {
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encCfg),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	return zap.New(core).Sugar()
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
