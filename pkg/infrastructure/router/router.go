package router

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Path of route
const (
	apiPath        = "/api"
	graphQLPath    = "/query"
	PlaygroundPath = "/playground"
)

const (
	QueryPath = apiPath + graphQLPath
)

// Options of router
type Options struct {
	Auth bool
}

// New creates route endpoint
func New(srv *handler.Server, options Options) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderXRequestedWith,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	}))

	e.GET("/health_check", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	{ // Apply RefreshAuth middleware only to the refresh token mutation endpoint
		e.POST(QueryPath, echo.WrapHandler(srv))
		e.GET(PlaygroundPath, func(c echo.Context) error {
			playground.Handler("GraphQL", QueryPath).ServeHTTP(c.Response(), c.Request())
			return nil
		})
	}

	return e
}
