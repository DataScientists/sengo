package profileentryrepository

import (
	"context"
	"sheng-go-backend/ent"
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
