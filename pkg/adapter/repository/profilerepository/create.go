package profilerepository

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
)

func (r *profileRepository) Create(
	ctx context.Context,
	input model.CreateProfileInput,
) (*model.Profile, error) {
	profile, err := r.client.Profile.Create().SetInput(input).Save(ctx)
	if err != nil {
		return nil, model.NewDBError(err)
	}
	return profile, nil
}

