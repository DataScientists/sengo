package profilefetcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sheng-go-backend/config"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/jobexecutionhistory"
	"sheng-go-backend/ent/profileentry"
	"sheng-go-backend/ent/schema/ulid"
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

	// Create a per-run log file in the project root
	logFile, fileLogger, err := createJobLogFile()
	if err != nil {
		pf.logger.Warnw("failed to create job log file, continuing with stdout only", "error", err)
	} else {
		defer logFile.Close()
		pf.logger = fileLogger
		pf.logger.Infof("Log file created: %s", logFile.Name())
	}

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
	var processedEntryIDs []ulid.ID
	var errors []string
	batchNumber := 0

	for {
		batchNumber++
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

				savedHistory, _ := pf.jobHistoryRepo.Create(ctx, history, nil)
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
		pf.logger.Infof("%s[%s] Fetching from DB: getting pending profile entries (batch size: %d)%s",
			colorCyan, time.Now().Format("2006-01-02 15:04:05"), allowedBatchSize, colorReset)
		pendingEntries, err := pf.profileEntryRepo.GetPendingBatch(ctx, allowedBatchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to get pending entries: %w", err)
		}
		pf.logger.Infof("%s[%s] Fetched from DB: total to process = %d, fetched = %d%s",
			colorGreen, time.Now().Format("2006-01-02 15:04:05"), allowedBatchSize, len(pendingEntries), colorReset)

		if len(pendingEntries) == 0 {
			pf.logger.Info("no pending profile entries, exiting")
			break
		}

		// Process each entry in the batch
		for i, entry := range pendingEntries {
			pf.logger.Infof("%s[%s] Processing entry %d/%d - URN: %s%s",
				colorCyan, time.Now().Format("2006-01-02 15:04:05"), i+1, len(pendingEntries), entry.LinkedinUrn, colorReset)
			// Update status to FETCHING
			pf.logger.Infof("%s[%s] Updating DB: setting status to FETCHING for entry %s%s",
				colorYellow, time.Now().Format("2006-01-02 15:04:05"), entry.LinkedinUrn, colorReset)
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
				pf.logger.Infof("%s[%s] Updating DB: setting status to FAILED for entry %s%s",
					colorRed, time.Now().Format("2006-01-02 15:04:05"), entry.LinkedinUrn, colorReset)
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
			pf.logger.Infof("%s[%s] Inserting/Updating DB: upserting profile for URN %s%s",
				colorYellow, time.Now().Format("2006-01-02 15:04:05"), entry.LinkedinUrn, colorReset)
			dbProfile := pf.convertToDBProfile(profile, rawS3Key, cleanedS3Key)
			if _, err := pf.profileRepo.Upsert(ctx, dbProfile); err != nil {
				errMsg := fmt.Sprintf("DB upsert failed: %v", err)
				pf.logger.Infof("%s[%s] Updating DB: setting status to FAILED for entry %s%s",
					colorRed, time.Now().Format("2006-01-02 15:04:05"), entry.LinkedinUrn, colorReset)
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
			pf.logger.Infof("%s[%s] DB operation complete: profile upserted successfully%s",
				colorGreen, time.Now().Format("2006-01-02 15:04:05"), colorReset)

			// Update profile entry as completed
			pf.logger.Infof("%s[%s] Updating DB: setting status to COMPLETED for entry %s%s",
				colorYellow, time.Now().Format("2006-01-02 15:04:05"), entry.LinkedinUrn, colorReset)
			if _, err := pf.profileEntryRepo.UpdateAfterFetch(ctx, string(entry.ID), rawS3Key, cleanedS3Key); err != nil {
				pf.logger.Warnw(
					"failed to update profile entry after fetch",
					"urn",
					entry.LinkedinUrn,
					"error",
					err,
				)
			}

			processedEntryIDs = append(processedEntryIDs, entry.ID)
			successCount++
			pf.logger.Infof("%s[%s] Batch #%d progress: %d/%d entries processed (success: %d, failed: %d)%s",
				colorGreen, time.Now().Format("2006-01-02 15:04:05"), batchNumber, i+1, len(pendingEntries), successCount, failedCount, colorReset)

			// Add 5s delay between profile fetch calls to avoid hitting RapidAPI rate limits
			if i < len(pendingEntries)-1 {
				pf.logger.Infof("%s[%s] Waiting 5 seconds before next fetch...%s",
					colorMagenta, time.Now().Format("2006-01-02 15:04:05"), colorReset)
				if err := sleepWithContext(ctx, 5*time.Second); err != nil {
					pf.logger.Warnf("%s[CANCELLED]%s Context cancelled during inter-fetch delay", colorRed, colorReset)
					break
				}
			}
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

	savedHistory, err := pf.jobHistoryRepo.Create(ctx, history, processedEntryIDs)
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

// ANSI color codes for logging
const (
	colorReset   = "\033[0m"
	colorCyan    = "\033[36m" // Fetching
	colorYellow  = "\033[33m" // Rate limit hit
	colorMagenta = "\033[35m" // Sleeping
	colorGreen   = "\033[32m" // Success
	colorRed     = "\033[31m" // Error/Failed
)

func (pf *ProfileFetcher) fetchProfileWithRetry(
	ctx context.Context,
	urn string,
) (*rapidapi.LinkedInProfile, []byte, int, error) {
	cfg := config.C.RapidAPI

	// maxRetries applies to non-rate-limit errors only
	maxRetries := cfg.RateLimitMaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	backoff := time.Duration(cfg.RateLimitBackoffMs) * time.Millisecond
	if backoff <= 0 {
		backoff = time.Second
	}

	// Cap max backoff at 60 seconds for rate limits
	maxBackoff := 60 * time.Second

	attempts := 0
	nonRateLimitAttempts := 0
	var lastErr error

	// Infinite loop - only exits on success, context cancellation, or non-rate-limit error after maxRetries
	for {
		attempts++

		// Log fetching attempt
		pf.logger.Infof("%s[FETCHING]%s URN: %s (attempt %d)", colorCyan, colorReset, urn, attempts)

		profile, rawData, err := pf.linkedinClient.FetchProfileByURN(ctx, urn)
		if err == nil {
			pf.logger.Infof("%s[SUCCESS]%s URN: %s fetched successfully after %d attempts", colorGreen, colorReset, urn, attempts)
			return profile, rawData, attempts, nil
		}

		lastErr = err

		// Check if this is a rate limit error
		var rateErr *rapidapi.RateLimitError
		if errors.As(err, &rateErr) {
			// Rate limit error - ALWAYS retry, never give up
			pf.logger.Warnf("%s[RATE LIMIT]%s URN: %s - API rate limit hit (attempt %d)", colorYellow, colorReset, urn, attempts)

			sleep := backoff
			if rateErr.RetryAfter > 0 && rateErr.RetryAfter > sleep {
				sleep = rateErr.RetryAfter
			}
			if sleep > maxBackoff {
				sleep = maxBackoff
			}

			pf.logger.Infof("%s[SLEEPING]%s URN: %s - waiting %v before retry...", colorMagenta, colorReset, urn, sleep)

			if err := sleepWithContext(ctx, sleep); err != nil {
				pf.logger.Warnf("%s[CANCELLED]%s URN: %s - context cancelled during sleep", colorRed, colorReset, urn)
				return nil, nil, attempts, err
			}

			// Exponential backoff, capped at maxBackoff
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			pf.logger.Infof("%s[RETRYING]%s URN: %s - resuming after rate limit", colorCyan, colorReset, urn)
			continue
		}

		// Non-rate-limit error - apply limited retries
		nonRateLimitAttempts++
		pf.logger.Errorf("%s[ERROR]%s URN: %s - non-rate-limit error: %v (attempt %d/%d)",
			colorRed, colorReset, urn, err, nonRateLimitAttempts, maxRetries)

		if nonRateLimitAttempts >= maxRetries {
			pf.logger.Errorf("%s[FAILED]%s URN: %s - giving up after %d non-rate-limit errors",
				colorRed, colorReset, urn, nonRateLimitAttempts)
			return nil, nil, attempts, lastErr
		}

		// Brief wait before retrying non-rate-limit errors
		pf.logger.Infof("%s[SLEEPING]%s URN: %s - waiting 1s before retry after error...", colorMagenta, colorReset, urn)
		if err := sleepWithContext(ctx, time.Second); err != nil {
			return nil, nil, attempts, err
		}
	}
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
		// Use Country/City if available, fallback to CountryName/CityName
		country := profile.Geo.Country
		if country == "" {
			country = profile.Geo.CountryName
		}
		city := profile.Geo.City
		if city == "" {
			city = profile.Geo.CityName
		}

		p.Country = &country
		p.City = &city
		p.GeoData = map[string]interface{}{
			"country_name": country,
			"city_name":    city,
			"full":         profile.Geo.Full,
			"country_code": profile.Geo.CountryCode,
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

// getProjectRoot returns the project root directory by finding the go.mod file
func getProjectRoot() string {
	// Use runtime.Caller to get the path of this source file, then walk up to find go.mod
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

// createJobLogFile creates a timestamped log file in the project root and returns
// a logger that writes to both stdout and the file.
func createJobLogFile() (*os.File, *zap.SugaredLogger, error) {
	projectRoot := getProjectRoot()
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFileName := fmt.Sprintf("profile_fetcher_%s.log", timestamp)
	logFilePath := filepath.Join(projectRoot, logFileName)

	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create log file %s: %w", logFilePath, err)
	}

	// Encoder config for the file (no color codes)
	fileEncCfg := zap.NewProductionEncoderConfig()
	fileEncCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	fileEncCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	// Encoder config for stdout (with color codes)
	stdoutEncCfg := zap.NewProductionEncoderConfig()
	stdoutEncCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	stdoutEncCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// Tee core: write to both stdout and the log file
	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(stdoutEncCfg),
			zapcore.AddSync(os.Stdout),
			zapcore.InfoLevel,
		),
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(fileEncCfg),
			zapcore.AddSync(logFile),
			zapcore.InfoLevel,
		),
	)

	logger := zap.New(core).Sugar()
	return logFile, logger, nil
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
