package profilerepository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/profile"
	ur "sheng-go-backend/pkg/usecase/repository"
)

// ProfileRepository interface for profile operations
type ProfileRepository interface {
	GetByURN(ctx context.Context, urn string) (*ent.Profile, error)
	Upsert(ctx context.Context, p *ent.Profile) (*ent.Profile, error)
}

type profileRepository struct {
	client *ent.Client
}

func NewProfileRepository(client *ent.Client) ur.Profile {
	return &profileRepository{client}
}

// NewProfileRepo creates a new ProfileRepository
func NewProfileRepo(client *ent.Client) ProfileRepository {
	return &profileRepository{client}
}

// GetByURN retrieves a profile by LinkedIn URN
func (r *profileRepository) GetByURN(ctx context.Context, urn string) (*ent.Profile, error) {
	return r.client.Profile.
		Query().
		Where(profile.Urn(urn)).
		Only(ctx)
}

// Upsert creates or updates a profile
func (r *profileRepository) Upsert(ctx context.Context, p *ent.Profile) (*ent.Profile, error) {
	// Try to find existing profile by URN
	existing, err := r.GetByURN(ctx, p.Urn)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	// If profile exists, update it
	if existing != nil {
		updateBuilder := r.client.Profile.UpdateOneID(existing.ID)

		if p.Username != nil {
			updateBuilder = updateBuilder.SetUsername(*p.Username)
		}
		if p.FirstName != nil {
			updateBuilder = updateBuilder.SetFirstName(*p.FirstName)
		}
		if p.LastName != nil {
			updateBuilder = updateBuilder.SetLastName(*p.LastName)
		}
		if p.Name != nil {
			updateBuilder = updateBuilder.SetName(*p.Name)
		}
		if p.Headline != nil {
			updateBuilder = updateBuilder.SetHeadline(*p.Headline)
		}
		if p.Title != nil {
			updateBuilder = updateBuilder.SetTitle(*p.Title)
		}
		if p.Country != nil {
			updateBuilder = updateBuilder.SetCountry(*p.Country)
		}
		if p.City != nil {
			updateBuilder = updateBuilder.SetCity(*p.City)
		}
		if p.Educations != nil {
			updateBuilder = updateBuilder.SetEducations(p.Educations)
		}
		if p.Positions != nil {
			updateBuilder = updateBuilder.SetPositions(p.Positions)
		}
		if p.Skills != nil {
			updateBuilder = updateBuilder.SetSkills(p.Skills)
		}
		if p.GeoData != nil {
			updateBuilder = updateBuilder.SetGeoData(p.GeoData)
		}
		if p.RawDataS3Key != nil {
			updateBuilder = updateBuilder.SetRawDataS3Key(*p.RawDataS3Key)
		}
		if p.CleanedDataS3Key != nil {
			updateBuilder = updateBuilder.SetCleanedDataS3Key(*p.CleanedDataS3Key)
		}

		return updateBuilder.Save(ctx)
	}

	// Create new profile
	createBuilder := r.client.Profile.
		Create().
		SetUrn(p.Urn)

	if p.Username != nil {
		createBuilder = createBuilder.SetUsername(*p.Username)
	}
	if p.FirstName != nil {
		createBuilder = createBuilder.SetFirstName(*p.FirstName)
	}
	if p.LastName != nil {
		createBuilder = createBuilder.SetLastName(*p.LastName)
	}
	if p.Name != nil {
		createBuilder = createBuilder.SetName(*p.Name)
	}
	if p.Headline != nil {
		createBuilder = createBuilder.SetHeadline(*p.Headline)
	}
	if p.Title != nil {
		createBuilder = createBuilder.SetTitle(*p.Title)
	}
	if p.Country != nil {
		createBuilder = createBuilder.SetCountry(*p.Country)
	}
	if p.City != nil {
		createBuilder = createBuilder.SetCity(*p.City)
	}
	if p.Educations != nil {
		createBuilder = createBuilder.SetEducations(p.Educations)
	}
	if p.Positions != nil {
		createBuilder = createBuilder.SetPositions(p.Positions)
	}
	if p.Skills != nil {
		createBuilder = createBuilder.SetSkills(p.Skills)
	}
	if p.GeoData != nil {
		createBuilder = createBuilder.SetGeoData(p.GeoData)
	}
	if p.RawDataS3Key != nil {
		createBuilder = createBuilder.SetRawDataS3Key(*p.RawDataS3Key)
	}
	if p.CleanedDataS3Key != nil {
		createBuilder = createBuilder.SetCleanedDataS3Key(*p.CleanedDataS3Key)
	}

	return createBuilder.Save(ctx)
}
