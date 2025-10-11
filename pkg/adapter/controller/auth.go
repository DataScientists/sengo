package controller

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
	usecase "sheng-go-backend/pkg/usecase/usecase/auth"
)

type Auth interface {
	Login(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error)
	RefreshToken(ctx context.Context) (*model.RefreshTokenPayload, error)
}

type authController struct {
	authUseCase usecase.Auth
}

func NewAuthController(au usecase.Auth) Auth {
	return &authController{authUseCase: au}
}

func (au *authController) Login(
	ctx context.Context,
	input model.LoginInput,
) (*model.AuthPayload, error) {
	return au.authUseCase.Login(ctx, input)
}

func (au *authController) RefreshToken(
	ctx context.Context,
) (*model.RefreshTokenPayload, error) {
	return au.authUseCase.RefreshToken(ctx)
}
