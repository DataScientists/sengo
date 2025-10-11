package userrepository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/schema/ulid"
	"sheng-go-backend/ent/user"
	"sheng-go-backend/pkg/entity/model"
)

func (r *userRepository) Get(ctx context.Context, id *ulid.ID) (*model.User, error) {
	u, err := r.client.User.Query().Where(user.IDEQ(*id)).Only(ctx)
	if err != nil {
		return nil, model.NewDBError(err)
	}
	return u, nil
}

func (r *userRepository) List(
	ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int,
	where *model.UserWhereInput,
) (*model.UserConnection, error) {
	uc, err := r.client.User.Query().
		Paginate(ctx, after, first, before, last, ent.WithUserFilter(where.Filter))
	if err != nil {
		return nil, model.NewDBError(err)
	}

	return uc, nil
}
