//go:generate mockgen -source=user.go -destination=./mocks/user_repository_mock.go -package=mocks
package repository

import (
	"context"
	"sheng-go-backend/ent/schema/ulid"
	"sheng-go-backend/pkg/entity/model"
)

// Repository Interface

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
