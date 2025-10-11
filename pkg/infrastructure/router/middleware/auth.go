package middleware

import (
	"errors"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/infrastructure/router/handler"
	"sheng-go-backend/pkg/util/auth"

	"github.com/labstack/echo/v4"
)

// AuthOptions of options for auth
type AuthOptions struct {
	Skip bool
}

// Auth is a middleware of authenticating users
func Auth(opts AuthOptions) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if opts.Skip {
				return next(c)
			}

			ctx := c.Request().Context()

			header := c.Request().Header.Get(echo.HeaderAuthorization)

			if header == "" {
				return handler.HandleError(c, model.NewAuthError(errors.New("Missing jwt token")))
			}
			accessToken, err := auth.GetTokenFromBearer(header)
			if err != nil {
				return handler.HandleError(c, model.NewAuthError(err))
			}

			ctx, err = auth.SetTokenToContext(ctx, accessToken)
			if err != nil {
				return handler.HandleError(c, model.NewAuthError(err))
			}

			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
