package controller

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
	usecase "sheng-go-backend/pkg/usecase/usecase/profileentry"
)

type ProfileEntry interface {
	Get(ctx context.Context, id *model.ID) (*model.ProfileEntry, error)
	Create(ctx context.Context, input model.CreateProfileEntryInput) (*model.ProfileEntry, error)
	Update(ctx context.Context, input model.UpdateProfileEntryInput) (*model.ProfileEntry, error)
	List(ctx context.Context,
		after *model.Cursor,
		first *int,
		before *model.Cursor,
		last *int, where *model.ProfileEntryWhereInput) (*model.ProfileEntryConnection, error)
}

type profileEntryController struct {
	profileEntryUseCase usecase.ProfileEntry
}

func NewProfileEntryController(pu usecase.ProfileEntry) ProfileEntry {
	return &profileEntryController{profileEntryUseCase: pu}
}

func (pc *profileEntryController) Get(
	ctx context.Context,
	id *model.ID,
) (*model.ProfileEntry, error) {
	return pc.profileEntryUseCase.Get(ctx, id)
}

func (pc *profileEntryController) Create(
	ctx context.Context,
	input model.CreateProfileEntryInput,
) (*model.ProfileEntry, error) {
	return pc.profileEntryUseCase.Create(ctx, input)
}

func (pc *profileEntryController) Update(
	ctx context.Context,
	input model.UpdateProfileEntryInput,
) (*model.ProfileEntry, error) {
	return pc.profileEntryUseCase.Update(ctx, input)
}

func (pc *profileEntryController) List(
	ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int, where *model.ProfileEntryWhereInput,
) (*model.ProfileEntryConnection, error) {
	return pc.profileEntryUseCase.List(ctx, after, first, before, last, where)
}
