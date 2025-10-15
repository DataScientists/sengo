package usecase

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/usecase/repository"
)

type profileUseCase struct {
	profileRepository repository.Profile
}

type Profile interface {
	Get(ctx context.Context, where *model.ProfileWhereInput) (*model.Profile, error)
	Create(ctx context.Context, input model.CreateProfileInput) (*model.Profile, error)
	Update(ctx context.Context, input model.UpdateProfileInput) (*model.Profile, error)
	List(ctx context.Context,
		after *model.Cursor,
		first *int,
		before *model.Cursor,
		last *int, where *model.ProfileWhereInput) (*model.ProfileConnection, error)
	GroupByTitle(
		ctx context.Context,
		searchTerm *string,
		minCount int,
	) ([]*model.ProfileTitleGroup, error)
}

func NewProfileUseCase(r repository.Profile) Profile {
	return &profileUseCase{profileRepository: r}
}

func (p *profileUseCase) Get(
	ctx context.Context,
	where *model.ProfileWhereInput,
) (*model.Profile, error) {
	return p.profileRepository.Get(ctx, where)
}

func (p *profileUseCase) Create(
	ctx context.Context,
	input model.CreateProfileInput,
) (*model.Profile, error) {
	return p.profileRepository.Create(ctx, input)
}

func (p *profileUseCase) Update(
	ctx context.Context,
	input model.UpdateProfileInput,
) (*model.Profile, error) {
	return p.profileRepository.Update(ctx, input)
}

func (p *profileUseCase) List(
	ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int, where *model.ProfileWhereInput,
) (*model.ProfileConnection, error) {
	return p.profileRepository.List(ctx, after, first, before, last, where)
}

func (p *profileUseCase) GroupByTitle(
	ctx context.Context,
	searchTerm *string,
	minCount int,
) ([]*model.ProfileTitleGroup, error) {
	return p.profileRepository.GroupByTitle(ctx, searchTerm, minCount)
}
