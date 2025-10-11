package usecase

import (
	"context"
	"errors"
	"fmt"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/usecase/repository"
	"sheng-go-backend/pkg/util/auth"
)

type authUseCase struct {
	authRepository repository.Auth
}

type Auth interface {
	Login(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error)
	RefreshToken(ctx context.Context) (*model.RefreshTokenPayload, error)
}

// This function creates new auth usercase

func NewAuthUseCase(r repository.Auth) Auth {
	return &authUseCase{authRepository: r}
}

func (uc *authUseCase) Login(
	ctx context.Context,
	input model.LoginInput,
) (*model.AuthPayload, error) {
	// Fetch User by email
	if err := ValidateLoginInput(input); err != nil {
		return nil, err
	}
	user, err := uc.authRepository.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, errors.New("Invalid Credentials")
	}
	err = auth.VerifyPassword(input.Password, user.Password)

	fmt.Printf("input: %+v\n", input)
	fmt.Printf("err: %+v\n", err)
	if err != nil {
		return nil, errors.New("Invalid Credentials")
	}
	accessToken, err := auth.GenerateAccessToken(string(user.ID))
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken(string(user.ID))
	if err != nil {
		return nil, err
	}
	return &model.AuthPayload{
		AccessToken:  accessToken,
		User:         user,
		RefreshToken: refreshToken,
	}, nil
}

func (uc *authUseCase) RefreshToken(ctx context.Context) (*model.RefreshTokenPayload, error) {
	refreshToken, err := auth.GetTokenRefreshFromContext(ctx)
	if err != nil {
		return nil, err
	}
	accessToken, newRefreshToken, err := auth.RefreshTokens(refreshToken)
	if err != nil {
		return nil, err
	}

	return &model.RefreshTokenPayload{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
