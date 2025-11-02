//go:generate mockgen -source=profile.go -destination=./mocks/profile_repository_mock.go -package=mocks
package repository

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
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
