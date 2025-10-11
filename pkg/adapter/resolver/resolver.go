package resolver

import (
	"sheng-go-backend/ent"
	"sheng-go-backend/graph/generated"
	"sheng-go-backend/pkg/adapter/controller"
	"sheng-go-backend/pkg/adapter/resolver/directives"

	"github.com/99designs/gqlgen/graphql"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	client     *ent.Client
	controller controller.Controller
}

// New schema creates NewExecutable Schema
func NewSchema(client *ent.Client, controller controller.Controller) graphql.ExecutableSchema {
	return generated.NewExecutableSchema(generated.Config{
		Resolvers: &Resolver{
			client:     client,
			controller: controller,
		},
		Directives: generated.DirectiveRoot{
			RefreshToken: directives.RefreshTokenDirective,
			Auth:         directives.AuthDirective,
		},
	})
}
