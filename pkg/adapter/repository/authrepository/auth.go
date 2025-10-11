package authrepository

import (
	"context"
	"errors"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/user"
	"sheng-go-backend/pkg/entity/model"
	repoInterface "sheng-go-backend/pkg/usecase/repository"
)

type authRepository struct {
	client *ent.Client
}

func NewAuthRepository(client *ent.Client) repoInterface.Auth {
	return &authRepository{client: client}
}

func (ar *authRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	res, err := ar.client.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err != nil {
		if ent.IsNotSingular(err) {
			return nil, model.NewNotFoundError(err, nil)
		}
		if ent.IsNotFound(err) {
			return nil, model.NewDBError(errors.New("User not found"))
		}
		return nil, model.NewDBError(err)
	}

	return res, nil
}
