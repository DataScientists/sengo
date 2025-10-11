package userrepository_test

import (
	"context"
	"sheng-go-backend/pkg/adapter/repository/userrepository"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserRepository_Create(t *testing.T) {
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
		arrange  func(t *testing.T)
		act      func(ctx context.Context, t *testing.T) (u *model.User, err error)
		assert   func(t *testing.T, u *model.User, err error)
		args     args
		teardown func(t *testing.T)
	}{
		{
			name: "It should create user",
			arrange: func(t *testing.T) {
				ctx := context.Background()
				_, err := client.User.Delete().Exec(ctx)
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
			},
			act: func(ctx context.Context, t *testing.T) (u *model.User, err error) {
				userInput := &model.CreateUserInput{
					Name:     "test",
					Age:      34,
					Password: "passwordd",
					Email:    "test@mail.com",
				}
				return repo.Create(ctx, *userInput)
			},

			assert: func(t *testing.T, u *model.User, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, u)
				assert.Equal(t, "test", u.Name)
				assert.Equal(t, 34, u.Age)
				assert.NotNil(t, u.Password)
				assert.Equal(t, "test@mail.com", u.Email)
				assert.Equal(t, 34, u.Age)
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
			tt.arrange(t)
			// Act

			got, err := tt.act(tt.args.ctx, t)

			// Assert

			tt.assert(t, got, err)

			// TearDown
			tt.teardown(t)
		})
	}
}
