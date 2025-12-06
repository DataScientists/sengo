package profileentryrepository

import (
	"context"
	"fmt"
	"sheng-go-backend/ent/profile"
	"sheng-go-backend/ent/profileentry"
	"sheng-go-backend/pkg/entity/model"
)

func (r *profileentryRepository) Create(
	ctx context.Context,
	input model.CreateProfileEntryInput,
) (*model.ProfileEntry, error) {
	existsEntry, err := r.client.ProfileEntry.
		Query().
		Where(profileentry.LinkedinUrnEQ(input.LinkedinUrn)).
		Exist(ctx)
	if err != nil {
		return nil, model.NewDBError(err)
	}
	if existsEntry {
		return nil, model.NewValidationError(fmt.Errorf("profile entry already exists for urn %s", input.LinkedinUrn))
	}

	existsProfile, err := r.client.Profile.
		Query().
		Where(
			profile.Or(
				profile.UrnEQ(input.LinkedinUrn),
				profile.UsernameEQ(input.LinkedinUrn),
			),
		).
		Exist(ctx)
	if err != nil {
		return nil, model.NewDBError(err)
	}
	if existsProfile {
		return nil, model.NewValidationError(fmt.Errorf("profile already exists for urn/username %s", input.LinkedinUrn))
	}

	profile, err := r.client.ProfileEntry.Create().SetInput(input).Save(ctx)
	if err != nil {
		return nil, model.NewDBError(err)
	}
	return profile, nil
}
