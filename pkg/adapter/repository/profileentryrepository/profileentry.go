package profileentryrepository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/profileentry"
	"sheng-go-backend/ent/schema/ulid"
	"sheng-go-backend/pkg/entity/model"
	ur "sheng-go-backend/pkg/usecase/repository"
	"time"
)

// ProfileEntryRepository interface for profile entry operations
type ProfileEntryRepository interface {
	GetPendingBatch(ctx context.Context, limit int) ([]*ent.ProfileEntry, error)
	UpdateStatus(
		ctx context.Context,
		id string,
		status profileentry.Status,
		errorMsg *string,
	) (*ent.ProfileEntry, error)
	UpdateAfterFetch(
		ctx context.Context,
		id string,
		rawS3Key, cleanedS3Key string,
	) (*ent.ProfileEntry, error)
	IncrementFetchCount(ctx context.Context, id string) error
	GetByStatus(
		ctx context.Context,
		status profileentry.Status,
		limit int,
	) ([]*ent.ProfileEntry, error)
	CountByStatus(ctx context.Context, status profileentry.Status) (int, error)
	GetStats(ctx context.Context) (*model.ProfileEntryStats, error)
	GetById(ctx context.Context, id ulid.ID) (*model.ProfileEntry, error)
}

type profileentryRepository struct {
	client *ent.Client
}

func NewprofileentryRepository(client *ent.Client) ur.ProfileEntry {
	return &profileentryRepository{client}
}

// NewProfileEntryRepository creates a new ProfileEntryRepository
func NewProfileEntryRepository(client *ent.Client) ProfileEntryRepository {
	return &profileentryRepository{client}
}

// GetPendingBatch retrieves a batch of pending profile entries
func (r *profileentryRepository) GetPendingBatch(
	ctx context.Context,
	limit int,
) ([]*ent.ProfileEntry, error) {
	return r.client.ProfileEntry.
		Query().
		Where(profileentry.StatusEQ(profileentry.StatusPending)).
		Order(ent.Asc(profileentry.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}

// UpdateStatus updates the status of a profile entry
func (r *profileentryRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status profileentry.Status,
	errorMsg *string,
) (*ent.ProfileEntry, error) {
	updateBuilder := r.client.ProfileEntry.
		UpdateOneID(ulid.ID(id)).
		SetStatus(status).
		SetUpdatedAt(time.Now())

	if errorMsg != nil {
		updateBuilder = updateBuilder.SetErrorMessage(*errorMsg)
	}

	return updateBuilder.Save(ctx)
}

// UpdateAfterFetch updates profile entry after successful fetch
func (r *profileentryRepository) UpdateAfterFetch(
	ctx context.Context,
	id string,
	rawS3Key, cleanedS3Key string,
) (*ent.ProfileEntry, error) {
	entry := r.client.ProfileEntry.Query().Where(profileentry.ID(ulid.ID(id))).OnlyX(ctx)
	return r.client.ProfileEntry.
		UpdateOneID(ulid.ID(id)).
		SetStatus(profileentry.StatusCOMPLETED).
		SetRawResponseS3Key(rawS3Key).
		SetTemplateJSONS3Key(cleanedS3Key).
		SetFetchCount(entry.FetchCount + 1).
		SetLastFetchedAt(time.Now()).
		Save(ctx)
}

// IncrementFetchCount increments the fetch count
func (r *profileentryRepository) IncrementFetchCount(ctx context.Context, id string) error {
	entry, err := r.client.ProfileEntry.Get(ctx, ulid.ID(id))
	if err != nil {
		return err
	}

	_, err = r.client.ProfileEntry.
		UpdateOneID(ulid.ID(id)).
		SetFetchCount(entry.FetchCount + 1).
		SetLastFetchedAt(time.Now()).
		Save(ctx)

	return err
}

// GetByStatus retrieves profile entries by status
func (r *profileentryRepository) GetByStatus(
	ctx context.Context,
	status profileentry.Status,
	limit int,
) ([]*ent.ProfileEntry, error) {
	query := r.client.ProfileEntry.
		Query().
		Where(profileentry.StatusEQ(status))

	if limit > 0 {
		query = query.Limit(limit)
	}

	return query.All(ctx)
}

// CountByStatus counts profile entries by status
func (r *profileentryRepository) CountByStatus(
	ctx context.Context,
	status profileentry.Status,
) (int, error) {
	return r.client.ProfileEntry.
		Query().
		Where(profileentry.StatusEQ(status)).
		Count(ctx)
}

// Get By Id
func (r *profileentryRepository) GetById(
	ctx context.Context,
	id model.ID,
) (*model.ProfileEntry, error) {
	u, err := r.client.ProfileEntry.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return u, nil
}

