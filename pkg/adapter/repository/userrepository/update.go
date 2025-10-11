package userrepository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/entity/model"
)

func (r *userRepository) Update(
	ctx context.Context,
	input model.UpdateUserInput,
) (*model.User, error) {
	u, err := r.client.User.UpdateOneID(input.ID).SetInput(input).Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, model.NewNotFoundError(err, input.ID)
		}

		return nil, model.NewDBError(err)
	}

	return u, nil
}
