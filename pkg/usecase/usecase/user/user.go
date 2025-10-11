package usecase

import (
	"context"
	"errors"
	"sheng-go-backend/ent/schema/ulid"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/usecase/repository"
	"sheng-go-backend/pkg/util/auth"
)

type userUseCase struct {
	userRepository repository.User
}

// User of usecase
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

// This function creates new user usercase
func NewUserUseCase(r repository.User) User {
	return &userUseCase{userRepository: r}
}

func (u *userUseCase) Get(ctx context.Context, id *ulid.ID) (*model.User, error) {
	if id == nil || *id == "" {
		return nil, errors.New("ID is missing")
	}
	return u.userRepository.Get(ctx, id)
}

func (u *userUseCase) Create(
	ctx context.Context,
	input model.CreateUserInput,
) (*model.User, error) {
	if err := ValidateCreateUserInput(input); err != nil {
		return nil, err
	}
	// Hash the password before createing the user
	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	input.Password = hashedPassword

	return u.userRepository.Create(ctx, input)
}

func (u *userUseCase) List(ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int,
	where *model.UserWhereInput,
) (*model.UserConnection, error) {
	return u.userRepository.List(ctx, after, first, before, last, where)
}

func (u *userUseCase) Update(
	ctx context.Context,
	input model.UpdateUserInput,
) (*model.User, error) {
	if err := ValidateUpdateUserInput(input); err != nil {
		return nil, err
	}
	return u.userRepository.Update(ctx, input)
}
