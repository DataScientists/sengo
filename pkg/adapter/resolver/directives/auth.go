package directives

import (
	"context"
	"errors"
	"sheng-go-backend/pkg/adapter/handler"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/util/auth"

	"github.com/99designs/gqlgen/graphql"
	"github.com/labstack/echo/v4"
)

func AuthDirective(
	ctx context.Context,
	obj interface{},
	next graphql.Resolver,
) (res interface{}, err error) {
	// Extract access token from the request headers
	reqCtx := graphql.GetOperationContext(ctx)
	headers := reqCtx.Headers
	token := headers.Get(echo.HeaderAuthorization)
	if token == "" {
		return nil, handler.HandleGraphQLError(
			ctx,
			model.NewAuthError(errors.New("jwt token is missing")),
		)
	}
	accessToken, err := auth.GetTokenFromBearer(token)
	if err != nil {
		return nil, handler.HandleGraphQLError(ctx, model.NewAuthError(err))
	}

	isExpired, err := auth.IsJWTExpired(accessToken)
	if err != nil {
		return nil, handler.HandleGraphQLError(ctx, err)
	}
	if isExpired {
		return nil, handler.HandleGraphQLError(ctx, model.NewAuthError(errors.New("Token Expired")))
	}

	ctx, err = auth.SetTokenToContext(ctx, accessToken)
	if err != nil {
		return nil, handler.HandleGraphQLError(ctx, model.NewAuthError(err))
	}

	// Call the next resolver
	return next(ctx)
}
