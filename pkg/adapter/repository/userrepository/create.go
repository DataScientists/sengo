package userrepository

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
)

func (r *userRepository) Create(
	ctx context.Context,
	input model.CreateUserInput,
) (*model.User, error) {
	u, err := r.client.User.Create().SetInput(input).Save(ctx)
	if err != nil {
		return nil, model.NewDBError(err)
	}

	return u, nil
}
