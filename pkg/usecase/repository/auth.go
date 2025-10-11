//go:generate mockgen -source=auth.go -destination=./mocks/auth_repository_mock.go -package=mocks

package repository

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
)

type Auth interface {
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
}
