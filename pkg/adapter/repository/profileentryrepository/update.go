package profileentryrepository

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
)

func (r *profileentryRepository) Update(
	ctx context.Context,
	input model.UpdateProfileEntryInput,
) (*model.ProfileEntry, error) {
	profile, err := r.client.ProfileEntry.UpdateOneID(input.ID).SetInput(input).Save(ctx)
	if err != nil {
		return nil, model.NewDBError(err)
	}
	return profile, nil
}
