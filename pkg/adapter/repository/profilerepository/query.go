package profilerepository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/entity/model"
)

func (r *profileRepository) Get(
	ctx context.Context,
	id model.ID,
) (*model.Profile, error) {
	res, err := r.client.Profile.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, model.NewDBError(err)
	}

	return res, nil
}

func (r *profileRepository) List(
	ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int,
	where *model.ProfileWhereInput,
) (*model.ProfileConnection, error) {
	pc, err := r.client.Profile.Query().
		Paginate(ctx, after, first, before, last, ent.WithProfileFilter(where.Filter))
	if err != nil {
		return nil, model.NewDBError(err)
	}

	return pc, nil
}
