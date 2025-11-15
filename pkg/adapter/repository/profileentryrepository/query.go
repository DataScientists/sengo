package profileentryrepository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/profileentry"
	"sheng-go-backend/pkg/entity/model"
)

func (r *profileentryRepository) Get(
	ctx context.Context,
	id *model.ID,
) (*model.ProfileEntry, error) {
	u, err := r.client.ProfileEntry.Get(ctx, *id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, model.NewNotFoundError(err, id)
		}
		return nil, model.NewDBError(err)
	}
	return u, nil
}

func (r *profileentryRepository) List(
	ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int,
	where *model.ProfileEntryWhereInput,
) (*model.ProfileEntryConnection, error) {
	pc, err := r.client.ProfileEntry.Query().
		Paginate(ctx, after, first, before, last, ent.WithProfileEntryFilter(where.Filter))
	if err != nil {
		return nil, model.NewDBError(err)
	}

	return pc, nil
}

func (r *profileentryRepository) Delete(
	ctx context.Context,
	id *model.ID,
) error {
	err := r.client.ProfileEntry.DeleteOneID(*id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return model.NewNotFoundError(err, id)
		}
		return model.NewDBError(err)
	}
	return nil
}

func (r *profileentryRepository) GetStats(
	ctx context.Context,
) (*model.ProfileEntryStats, error) {
	// Get total count
	totalCount, err := r.client.ProfileEntry.Query().Count(ctx)
	if err != nil {
		return nil, model.NewDBError(err)
	}

	// Count by status
	pendingCount, err := r.CountByStatus(ctx, profileentry.StatusPending)
	if err != nil {
		return nil, model.NewDBError(err)
	}

	fetchingCount, err := r.CountByStatus(ctx, profileentry.StatusFetching)
	if err != nil {
		return nil, model.NewDBError(err)
	}

	completedCount, err := r.CountByStatus(ctx, profileentry.StatusCOMPLETED)
	if err != nil {
		return nil, model.NewDBError(err)
	}

	failedCount, err := r.CountByStatus(ctx, profileentry.StatusFAILED)
	if err != nil {
		return nil, model.NewDBError(err)
	}

	return &model.ProfileEntryStats{
		TotalCount:     totalCount,
		PendingCount:   pendingCount,
		FetchingCount:  fetchingCount,
		CompletedCount: completedCount,
		FailedCount:    failedCount,
	}, nil
}
