package apiquotatrackerrepository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/apiquotatracker"
	"sheng-go-backend/ent/schema/ulid"
	"time"
)

type APIQuotaTrackerRepository struct {
	client *ent.Client
}

func NewAPIQuotaTrackerRepository(client *ent.Client) *APIQuotaTrackerRepository {
	return &APIQuotaTrackerRepository{client: client}
}

// GetByMonthYear retrieves the quota tracker for a specific month/year
func (r *APIQuotaTrackerRepository) GetByMonthYear(ctx context.Context, month, year int) (*ent.APIQuotaTracker, error) {
	return r.client.APIQuotaTracker.
		Query().
		Where(
			apiquotatracker.Month(month),
			apiquotatracker.Year(year),
		).
		Only(ctx)
}

// GetCurrent retrieves the quota tracker for the current month
func (r *APIQuotaTrackerRepository) GetCurrent(ctx context.Context) (*ent.APIQuotaTracker, error) {
	now := time.Now()
	return r.GetByMonthYear(ctx, int(now.Month()), now.Year())
}

// Create creates a new quota tracker
func (r *APIQuotaTrackerRepository) Create(ctx context.Context, month, year, quotaLimit int) (*ent.APIQuotaTracker, error) {
	return r.client.APIQuotaTracker.
		Create().
		SetMonth(month).
		SetYear(year).
		SetCallCount(0).
		SetQuotaLimit(quotaLimit).
		SetQuotaExceeded(false).
		SetOverrideEnabled(false).
		SetNotificationSent(false).
		Save(ctx)
}

// IncrementCallCount increments the API call count
func (r *APIQuotaTrackerRepository) IncrementCallCount(ctx context.Context, id string, count int) (*ent.APIQuotaTracker, error) {
	tracker, err := r.client.APIQuotaTracker.Get(ctx, ulid.ID(id))
	if err != nil {
		return nil, err
	}

	newCount := tracker.CallCount + count
	quotaExceeded := newCount >= tracker.QuotaLimit

	return r.client.APIQuotaTracker.
		UpdateOneID(ulid.ID(id)).
		SetCallCount(newCount).
		SetQuotaExceeded(quotaExceeded).
		SetLastCallAt(time.Now()).
		Save(ctx)
}

// SetOverride sets the quota override flag
func (r *APIQuotaTrackerRepository) SetOverride(ctx context.Context, id string, enabled bool) (*ent.APIQuotaTracker, error) {
	return r.client.APIQuotaTracker.
		UpdateOneID(ulid.ID(id)).
		SetOverrideEnabled(enabled).
		Save(ctx)
}

// SetNotificationSent marks that quota exceeded notification has been sent
func (r *APIQuotaTrackerRepository) SetNotificationSent(ctx context.Context, id string, sent bool) (*ent.APIQuotaTracker, error) {
	return r.client.APIQuotaTracker.
		UpdateOneID(ulid.ID(id)).
		SetNotificationSent(sent).
		Save(ctx)
}

// UpdateQuotaLimit updates the monthly quota limit
func (r *APIQuotaTrackerRepository) UpdateQuotaLimit(ctx context.Context, id string, newLimit int) (*ent.APIQuotaTracker, error) {
	return r.client.APIQuotaTracker.
		UpdateOneID(ulid.ID(id)).
		SetQuotaLimit(newLimit).
		Save(ctx)
}

// ListHistory retrieves quota history for past months
func (r *APIQuotaTrackerRepository) ListHistory(ctx context.Context, limit int) ([]*ent.APIQuotaTracker, error) {
	return r.client.APIQuotaTracker.
		Query().
		Order(ent.Desc(apiquotatracker.FieldYear), ent.Desc(apiquotatracker.FieldMonth)).
		Limit(limit).
		All(ctx)
}
