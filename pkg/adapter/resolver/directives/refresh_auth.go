package directives

import (
	"context"
	"errors"
	"sheng-go-backend/pkg/adapter/handler"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/util/auth"

	"github.com/99designs/gqlgen/graphql"
)

func RefreshTokenDirective(
	ctx context.Context,
	obj interface{},
	next graphql.Resolver,
) (res interface{}, err error) {
	// Extract refresh token from the request headers
	reqCtx := graphql.GetOperationContext(ctx)
	headers := reqCtx.Headers
	refreshToken := headers.Get("RefreshToken")
	if refreshToken == "" {
		return nil, handler.HandleGraphQLError(
			ctx,
			model.NewAuthError(errors.New("RefreshToken is missing")),
		)
	}

	// ✅ Inject refresh token into the context
	ctx, err = auth.SetTokenRefreshToContext(ctx, refreshToken)
	if err != nil {
		return nil, handler.HandleGraphQLError(ctx, model.NewAuthError(err))
	}

	// Call the next resolver
	return next(ctx)
}
