//go:generate mockgen -source=profile.go -destination=./mocks/profile_repository_mock.go -package=mocks
package repository

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
)

type Profile interface {
	Get(ctx context.Context, where *model.ProfileWhereInput) (*model.Profile, error)
	Create(ctx context.Context, input model.CreateProfileInput) (*model.Profile, error)
	Update(ctx context.Context, input model.UpdateProfileInput) (*model.Profile, error)
	List(ctx context.Context,
		after *model.Cursor,
		first *int,
		before *model.Cursor,
		last *int, where *model.ProfileWhereInput) (*model.ProfileConnection, error)
}

