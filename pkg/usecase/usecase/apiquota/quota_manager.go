package apiquota

import (
	"context"
	"fmt"
	"sheng-go-backend/config"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/adapter/repository/apiquotatrackerrepository"
	"sheng-go-backend/pkg/infrastructure/email"
	"time"
)

// QuotaManager handles API quota tracking and enforcement
type QuotaManager struct {
	repo         *apiquotatrackerrepository.APIQuotaTrackerRepository
	emailService *email.EmailService
}

// NewQuotaManager creates a new QuotaManager
func NewQuotaManager(
	repo *apiquotatrackerrepository.APIQuotaTrackerRepository,
	emailService *email.EmailService,
) *QuotaManager {
	return &QuotaManager{
		repo:         repo,
		emailService: emailService,
	}
}

// CheckAndReserveQuota checks if quota is available and reserves it for the batch
func (qm *QuotaManager) CheckAndReserveQuota(ctx context.Context, batchSize int) (int, error) {
	// Get or create current month tracker
	tracker, err := qm.getOrCreateCurrentTracker(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get quota tracker: %w", err)
	}

	// Check if quota is exceeded and override is not enabled
	if tracker.QuotaExceeded && !tracker.OverrideEnabled {
		// Send notification if not already sent
		if !tracker.NotificationSent {
			if err := qm.sendQuotaExceededNotification(ctx, tracker); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Failed to send quota exceeded notification: %v\n", err)
			}
		}
		return 0, fmt.Errorf("API quota exceeded (%d/%d). Will auto-resume on 1st of next month. Admin can enable override.",
			tracker.CallCount, tracker.QuotaLimit)
	}

	// Calculate remaining quota
	remaining := tracker.QuotaLimit - tracker.CallCount

	// Determine allowed batch size
	allowedBatchSize := batchSize
	if remaining < batchSize {
		allowedBatchSize = remaining
		if allowedBatchSize <= 0 {
			return 0, fmt.Errorf("insufficient API quota remaining")
		}
	}

	return allowedBatchSize, nil
}

// IncrementCallCount increments the API call count
func (qm *QuotaManager) IncrementCallCount(ctx context.Context, count int) error {
	tracker, err := qm.getOrCreateCurrentTracker(ctx)
	if err != nil {
		return fmt.Errorf("failed to get quota tracker: %w", err)
	}

	// Increment the call count
	updatedTracker, err := qm.repo.IncrementCallCount(ctx, string(tracker.ID), count)
	if err != nil {
		return fmt.Errorf("failed to increment call count: %w", err)
	}

	// Check if quota just got exceeded
	if updatedTracker.QuotaExceeded && !tracker.QuotaExceeded {
		// Just exceeded, send notification
		if err := qm.sendQuotaExceededNotification(ctx, updatedTracker); err != nil {
			fmt.Printf("Failed to send quota exceeded notification: %v\n", err)
		}
	}

	return nil
}

// SetQuotaOverride sets the quota override flag
func (qm *QuotaManager) SetQuotaOverride(ctx context.Context, enabled bool) error {
	tracker, err := qm.getOrCreateCurrentTracker(ctx)
	if err != nil {
		return fmt.Errorf("failed to get quota tracker: %w", err)
	}

	_, err = qm.repo.SetOverride(ctx, string(tracker.ID), enabled)
	if err != nil {
		return fmt.Errorf("failed to set quota override: %w", err)
	}

	// Send notification about override status change
	if err := qm.emailService.SendQuotaOverrideAlert(enabled); err != nil {
		fmt.Printf("Failed to send quota override notification: %v\n", err)
	}

	return nil
}

// ResetMonthlyQuota resets the quota for a new month
func (qm *QuotaManager) ResetMonthlyQuota(ctx context.Context) error {
	now := time.Now()
	quotaLimit := config.C.RapidAPI.MonthlyQuota
	if quotaLimit == 0 {
		quotaLimit = 50000 // Default
	}

	// Create new tracker for current month
	_, err := qm.repo.Create(ctx, int(now.Month()), now.Year(), quotaLimit)
	if err != nil {
		// Might already exist if called multiple times
		if !ent.IsConstraintError(err) {
			return fmt.Errorf("failed to create new quota tracker: %w", err)
		}
	}

	// Send reset notification
	if err := qm.emailService.SendQuotaResetNotification(int(now.Month()), now.Year(), quotaLimit); err != nil {
		fmt.Printf("Failed to send quota reset notification: %v\n", err)
	}

	return nil
}

// GetCurrentQuotaStatus retrieves the current quota status
func (qm *QuotaManager) GetCurrentQuotaStatus(ctx context.Context) (*ent.APIQuotaTracker, error) {
	return qm.getOrCreateCurrentTracker(ctx)
}

// getOrCreateCurrentTracker gets or creates the quota tracker for current month
func (qm *QuotaManager) getOrCreateCurrentTracker(ctx context.Context) (*ent.APIQuotaTracker, error) {
	tracker, err := qm.repo.GetCurrent(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Create new tracker for current month
			now := time.Now()
			quotaLimit := config.C.RapidAPI.MonthlyQuota
			if quotaLimit == 0 {
				quotaLimit = 50000
			}

			tracker, err = qm.repo.Create(ctx, int(now.Month()), now.Year(), quotaLimit)
			if err != nil {
				return nil, fmt.Errorf("failed to create quota tracker: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get quota tracker: %w", err)
		}
	}

	return tracker, nil
}

// GetHistory retrieves quota history for past months
func (qm *QuotaManager) GetHistory(ctx context.Context, limit int) ([]*ent.APIQuotaTracker, error) {
	history, err := qm.repo.ListHistory(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota history: %w", err)
	}
	return history, nil
}

// UpdateLimit updates the quota limit for the current month
func (qm *QuotaManager) UpdateLimit(ctx context.Context, newLimit int) (*ent.APIQuotaTracker, error) {
	tracker, err := qm.getOrCreateCurrentTracker(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota tracker: %w", err)
	}

	updatedTracker, err := qm.repo.UpdateQuotaLimit(ctx, string(tracker.ID), newLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to update quota limit: %w", err)
	}

	return updatedTracker, nil
}

// sendQuotaExceededNotification sends notification when quota is exceeded
func (qm *QuotaManager) sendQuotaExceededNotification(ctx context.Context, tracker *ent.APIQuotaTracker) error {
	// Send email
	if err := qm.emailService.SendQuotaExceededAlert(
		tracker.CallCount,
		tracker.QuotaLimit,
		tracker.Month,
		tracker.Year,
	); err != nil {
		return err
	}

	// Mark notification as sent
	_, err := qm.repo.SetNotificationSent(ctx, string(tracker.ID), true)
	return err
}
