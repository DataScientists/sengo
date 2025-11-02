package profileentryrepository

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
)

func (r *profileentryRepository) Create(
	ctx context.Context,
	input model.CreateProfileEntryInput,
) (*model.ProfileEntry, error) {
	profile, err := r.client.ProfileEntry.Create().SetInput(input).Save(ctx)
	if err != nil {
		return nil, model.NewDBError(err)
	}
	return profile, nil
}
