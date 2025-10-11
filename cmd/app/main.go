package main

import (
	"log"
	"sheng-go-backend/config"
	"sheng-go-backend/ent"
	_ "sheng-go-backend/ent/runtime"
	"sheng-go-backend/pkg/adapter/controller"
	"sheng-go-backend/pkg/infrastructure/datastore"
	"sheng-go-backend/pkg/infrastructure/graphql"
	"sheng-go-backend/pkg/infrastructure/router"
	"sheng-go-backend/pkg/registry"
)

func main() {
	config.ReadConfig(config.ReadConfigOption{})

	client := newDBClient()
	ctrl := newController(client)

	srv := graphql.NewServer(client, ctrl)

	e := router.New(srv, router.Options{
		Auth: false,
	})

	e.Logger.Fatal(e.Start(":" + config.C.Server.Address))
}

func newDBClient() *ent.Client {
	client, err := datastore.NewClient()
	if err != nil {
		log.Fatalf("Failed to open db connection: %v", err)
	}
	return client
}

func newController(client *ent.Client) controller.Controller {
	r := registry.New(client)
	return r.NewController()
}
