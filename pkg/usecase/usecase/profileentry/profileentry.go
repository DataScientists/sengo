package usecase

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/usecase/repository"
)

type profileUseCase struct {
	profileRepository repository.ProfileEntry
}

type ProfileEntry interface {
	Get(ctx context.Context, id *model.ID) (*model.ProfileEntry, error)
	Create(ctx context.Context, input model.CreateProfileEntryInput) (*model.ProfileEntry, error)
	Update(ctx context.Context, input model.UpdateProfileEntryInput) (*model.ProfileEntry, error)
	Delete(ctx context.Context, id *model.ID) error
	List(ctx context.Context,
		after *model.Cursor,
		first *int,
		before *model.Cursor,
		last *int, where *model.ProfileEntryWhereInput) (*model.ProfileEntryConnection, error)
	GetStats(ctx context.Context) (*model.ProfileEntryStats, error)
}

func NewProfileEntryUseCase(r repository.ProfileEntry) ProfileEntry {
	return &profileUseCase{profileRepository: r}
}

func (p *profileUseCase) Get(
	ctx context.Context,
	id *model.ID,
) (*model.ProfileEntry, error) {
	return p.profileRepository.Get(ctx, id)
}

func (p *profileUseCase) Create(
	ctx context.Context,
	input model.CreateProfileEntryInput,
) (*model.ProfileEntry, error) {
	return p.profileRepository.Create(ctx, input)
}

func (p *profileUseCase) Update(
	ctx context.Context,
	input model.UpdateProfileEntryInput,
) (*model.ProfileEntry, error) {
	return p.profileRepository.Update(ctx, input)
}

func (p *profileUseCase) List(
	ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int, where *model.ProfileEntryWhereInput,
) (*model.ProfileEntryConnection, error) {
	return p.profileRepository.List(ctx, after, first, before, last, where)
}

func (p *profileUseCase) Delete(
	ctx context.Context,
	id *model.ID,
) error {
	return p.profileRepository.Delete(ctx, id)
}

func (p *profileUseCase) GetStats(
	ctx context.Context,
) (*model.ProfileEntryStats, error) {
	return p.profileRepository.GetStats(ctx)
}
