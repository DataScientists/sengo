package userrepository_test

import (
	"context"
	"sheng-go-backend/ent"
	userrepository "sheng-go-backend/pkg/adapter/repository/userrepository"
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

func TestUserRepository_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping  unit test")
	}
	t.Helper()

	client, teardown := setup(t)
	defer teardown()

	repo := userrepository.NewUserRepository(client)

	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name     string
		arrange  func(t *testing.T) *model.User
		act      func(ctx context.Context, t *testing.T, user model.User) (u *model.User, err error)
		assert   func(t *testing.T, u *model.User, err error)
		args     args
		teardown func(t *testing.T)
	}{
		{
			name: "It should get user",
			arrange: func(t *testing.T) *model.User {
				ctx := context.Background()
				_, err := client.User.Delete().Exec(ctx)
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
				user := &model.User{
					Name:     "test",
					Age:      34,
					Password: "passwodsfsd",
					Email:    "test@mail.com",
				}

				createdUser, err := client.User.Create().
					SetName(user.Name).
					SetAge(user.Age).
					SetPassword(user.Password).
					SetEmail(user.Email).
					Save(ctx)
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
				return createdUser
			},
			act: func(ctx context.Context, t *testing.T, user model.User) (u *model.User, err error) {
				return repo.Get(ctx, &user.ID)
			},

			assert: func(t *testing.T, u *model.User, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, u)
				assert.Equal(t, "test", u.Name)
				assert.Equal(t, 34, u.Age)
				assert.Equal(t, "test@mail.com", u.Email)
			},
			args: args{
				ctx: context.Background(),
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
			user := tt.arrange(t)
			// Act

			got, err := tt.act(tt.args.ctx, t, *user)

			// Assert

			tt.assert(t, got, err)

			// TearDown
			tt.teardown(t)
		})
	}
}

func TestUserRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping  unit test")
	}
	t.Helper()

	client, teardown := setup(t)
	defer teardown()

	repo := userrepository.NewUserRepository(client)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		arrange func(t *testing.T)
		act     func(ctx context.Context, t *testing.T) (uc *model.UserConnection, err error)
		assert  func(t *testing.T, uc *model.UserConnection, err error)
		args    struct {
			ctx context.Context
		}
		teardown func(t *testing.T)
	}{
		{
			name: "It should get user's list",
			arrange: func(t *testing.T) {
				ctx := context.Background()
				_, err := client.User.Delete().Exec(ctx)
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
				users := []struct {
					name     string
					age      int
					password string
					email    string
				}{
					{
						name:     "test",
						age:      33,
						password: "passwssord",
						email:    "test1@gmail.com",
					},
					{
						name:     "test2",
						age:      31,
						password: "passwordsafd",
						email:    "test2@gmail.com",
					},
					{
						name:     "Jigme",
						age:      25,
						password: "passssword",
						email:    "test3@gmail.com",
					},
				}
				bulk := make([]*ent.UserCreate, len(users))
				for i, u := range users {
					bulk[i] = client.User.Create().
						SetName(u.name).
						SetAge(u.age).
						SetPassword(u.password).SetEmail(u.email)
				}
				_, err = client.User.CreateBulk(bulk...).Save(ctx)
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
			},
			act: func(ctx context.Context, t *testing.T) (uc *model.UserConnection, err error) {
				first := 5
				return repo.List(ctx, nil, &first, nil, nil, nil)
			},
			assert: func(t *testing.T, uc *model.UserConnection, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 3, len(uc.Edges))
			},
			args: args{
				ctx: context.Background(),
			},
			teardown: func(t *testing.T) {
				testutil.DropUser(t, client)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.arrange(t)
			uc, err := tt.act(tt.args.ctx, t)
			tt.assert(t, uc, err)
			tt.teardown(t)
		})
	}
}
