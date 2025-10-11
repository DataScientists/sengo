package authrepository_test

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/adapter/repository/authrepository"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) (client *ent.Client, teardown func()) {
	testutil.ReadConfig()
	c := testutil.NewDBClient(t)

	return c, func() {
		testutil.DropUser(t, c)
		defer c.Close()
	}
}

func TestAuthRepository_Login(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping  unit test")
	}
	t.Helper()

	client, teardown := setup(t)
	defer teardown()

	repo := authrepository.NewAuthRepository(client)

	type args struct {
		ctx   context.Context
		email string
	}

	tests := []struct {
		name     string
		arrange  func(t *testing.T)
		act      func(ctx context.Context, t *testing.T, email string) (u *model.User, err error)
		assert   func(t *testing.T, u *model.User, err error)
		args     args
		teardown func(t *testing.T)
	}{
		{
			name: "It should get user",
			arrange: func(t *testing.T) {
				ctx := context.Background()
				_, err := client.User.Delete().Exec(ctx)
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
				user := &model.User{
					Name:     "test",
					Email:    "test@mail.com",
					Age:      34,
					Password: "Passworddsf",
				}
				if err != nil {
					t.Error(err)
					t.FailNow()
				}

				_, err = client.User.Create().
					SetName(user.Name).
					SetAge(user.Age).
					SetPassword(user.Password).
					SetEmail(user.Email).
					Save(ctx)
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
			},
			act: func(ctx context.Context, t *testing.T, email string) (u *model.User, err error) {
				return repo.GetUserByEmail(ctx, email)
			},

			assert: func(t *testing.T, u *model.User, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, u)
				assert.Equal(t, "test", u.Name)
				assert.Equal(t, 34, u.Age)
			},
			args: args{
				ctx:   context.Background(),
				email: "test@mail.com",
			},
			teardown: func(t *testing.T) {
				testutil.DropUser(t, client)
			},
		},
		{
			name: "It should return error when email is not in record",
			arrange: func(t *testing.T) {
				ctx := context.Background()
				_, err := client.User.Delete().Exec(ctx)
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
				user := &model.User{
					Name:     "test",
					Email:    "test@mail.com",
					Age:      34,
					Password: "Passworddsf",
				}
				if err != nil {
					t.Error(err)
					t.FailNow()
				}

				_, err = client.User.Create().
					SetName(user.Name).
					SetAge(user.Age).
					SetPassword(user.Password).
					SetEmail(user.Email).
					Save(ctx)
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
			},
			act: func(ctx context.Context, t *testing.T, email string) (u *model.User, err error) {
				return repo.GetUserByEmail(ctx, email)
			},

			assert: func(t *testing.T, u *model.User, err error) {
				assert.NotNil(t, err)
				assert.Nil(t, u)
			},
			args: args{
				ctx:   context.Background(),
				email: "test2@mail.com",
			},
			teardown: func(t *testing.T) {
				testutil.DropUser(t, client)
			},
		},
	}
	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tt.arrange(t)
			// Act

			got, err := tt.act(tt.args.ctx, t, tt.args.email)

			// Assert

			tt.assert(t, got, err)

			// TearDown
			tt.teardown(t)
		})
	}
}
