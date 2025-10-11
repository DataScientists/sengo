package testutil

import (
	"context"
	"database/sql"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/enttest"
	"sheng-go-backend/pkg/infrastructure/datastore"
	"testing"

	"entgo.io/ent/dialect"
	pgx "github.com/jackc/pgx/v5/stdlib"
)

// init registers the pgx stdlib driver under the name "postgres" so that ent recognizes it.
func init() {
	// stdlib.GetDefaultDriver returns a driver that implements database/sql/driver.Driver.
	// We register it under the name "postgres", which is one of the supported dialects in ent.
	sql.Register("postgres", pgx.GetDefaultDriver())
}

// NewDBClient loads database for test.
func NewDBClient(t *testing.T) *ent.Client {
	client := datastore.NewDSN()

	return enttest.Open(t, dialect.Postgres, client)
}

func NewSqlListeDBClient(t *testing.T) *ent.Client {
	return enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
}

// DropAll drops all the data from database

func DropAll(t *testing.T, client *ent.Client) {
	t.Log("drop data from database")
	DropUser(t, client)
	DropTodo(t, client)
}

// DropUser drops data from users.
func DropUser(t *testing.T, client *ent.Client) {
	ctx := context.Background()
	_, err := client.User.Delete().Exec(ctx)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

// DropTodo drops all the data from todos.

func DropTodo(t *testing.T, client *ent.Client) {
	ctx := context.Background()
	_, err := client.Todo.Delete().Exec(ctx)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}
