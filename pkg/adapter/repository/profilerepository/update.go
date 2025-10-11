package profilerepository

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
)

func (r *profileRepository) Update(
	ctx context.Context,
	input model.UpdateProfileInput,
) (*model.Profile, error) {
	profile, err := r.client.Profile.UpdateOneID(input.ID).SetInput(input).Save(ctx)
	if err != nil {
		return nil, model.NewDBError(err)
	}
	return profile, nil
}

