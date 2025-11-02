package profilerepository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/entity/model"
)

func (r *profileRepository) Get(
	ctx context.Context,
	where *model.ProfileWhereInput,
) (*model.Profile, error) {
	q := r.client.Profile.Query()

	q, err := where.Filter(q)
	if err != nil {
		return nil, model.NewInvalidParamError(nil)
	}

	res, err := q.Only(ctx)
	if err != nil {
		if ent.IsNotSingular(err) {
			return nil, model.NewNotFoundError(err, nil)
		}
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
