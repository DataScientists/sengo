package datastore

import (
	"context"
	"fmt"
	"sheng-go-backend/config"
	"sheng-go-backend/ent"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func NewDSN() string {
	dsn := "postgres://" + config.C.Database.User + ":" + config.C.Database.Password + "@" + config.C.Database.Addr + ":" + config.C.Database.Port + "/" + config.C.Database.DBName + "?sslmode=disable"
	return dsn
}

// Create a new Ent client using pgxpool
func NewClient() (*ent.Client, error) {
	// Create pgx connection pool
	dsn := NewDSN()
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to create pool config: %w", err)
	}
	poolConfig.MaxConns = 20
	poolConfig.MinConns = 0
	poolConfig.MaxConnLifetime = time.Minute * 2
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	// Use stdlib to wrap pgxpool in database/sql compatibility
	sqlDB := stdlib.OpenDBFromPool(pool)

	// Wrap the sql.DB with Ent's SQL driver
	drv := sql.OpenDB(dialect.Postgres, sqlDB)

	// Create the Ent client
	client := ent.NewClient(ent.Driver(drv), ent.Debug())

	return client, nil
}
