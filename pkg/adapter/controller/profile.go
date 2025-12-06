package controller

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
	usecase "sheng-go-backend/pkg/usecase/usecase/profile"
)

type Profile interface {
	Get(ctx context.Context, id model.ID) (*model.Profile, error)
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

type profileController struct {
	profileUseCase usecase.Profile
}

func NewProfileController(pu usecase.Profile) Profile {
	return &profileController{profileUseCase: pu}
}

func (pc *profileController) Get(
	ctx context.Context,
	id model.ID,
) (*model.Profile, error) {
	return pc.profileUseCase.Get(ctx, id)
}

func (pc *profileController) Create(
	ctx context.Context,
	input model.CreateProfileInput,
) (*model.Profile, error) {
	return pc.profileUseCase.Create(ctx, input)
}

func (pc *profileController) Update(
	ctx context.Context,
	input model.UpdateProfileInput,
) (*model.Profile, error) {
	return pc.profileUseCase.Update(ctx, input)
}

func (pc *profileController) List(
	ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int, where *model.ProfileWhereInput,
) (*model.ProfileConnection, error) {
	return pc.profileUseCase.List(ctx, after, first, before, last, where)
}

func (pc *profileController) GroupByTitle(
	ctx context.Context,
	searchTerm *string,
	minCount int,
) ([]*model.ProfileTitleGroup, error) {
	return pc.profileUseCase.GroupByTitle(ctx, searchTerm, minCount)
}
