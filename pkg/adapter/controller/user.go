package controller

import (
	"context"
	"sheng-go-backend/ent/schema/ulid"
	"sheng-go-backend/pkg/entity/model"
	usecase "sheng-go-backend/pkg/usecase/usecase/user"
)

type User interface {
	Get(ctx context.Context, id *ulid.ID) (*model.User, error)
	Create(ctx context.Context, input model.CreateUserInput) (*model.User, error)
	Update(ctx context.Context, input model.UpdateUserInput) (*model.User, error)
	List(
		ctx context.Context,
		after *model.Cursor,
		first *int,
		before *model.Cursor,
		last *int,
		where *model.UserWhereInput,
	) (*model.UserConnection, error)
}

type userController struct {
	userUseCase usecase.User
}

// This function will create new controller for user
func NewUserController(uu usecase.User) User {
	return &userController{userUseCase: uu}
}

func (uc *userController) Get(ctx context.Context, id *ulid.ID) (*model.User, error) {
	return uc.userUseCase.Get(ctx, id)
}

func (uc *userController) Create(
	ctx context.Context,
	input model.CreateUserInput,
) (*model.User, error) {
	return uc.userUseCase.Create(ctx, input)
}

func (uc *userController) Update(
	ctx context.Context,
	input model.UpdateUserInput,
) (*model.User, error) {
	return uc.userUseCase.Update(ctx, input)
}

func (uc *userController) List(ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int,
	where *model.UserWhereInput,
) (*model.UserConnection, error) {
	return uc.userUseCase.List(ctx, after, first, before, last, where)
}
