package userrepository_test

import (
	"context"
	"sheng-go-backend/pkg/adapter/repository/userrepository"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserRepository_Update(t *testing.T) {
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
		arrange  func(t *testing.T) (u *model.User)
		act      func(ctx context.Context, t *testing.T, user model.User) (u *model.User, err error)
		assert   func(t *testing.T, user *model.User, err error)
		args     args
		teardown func(t *testing.T)
	}{
		{
			name: "It should update the user",
			arrange: func(t *testing.T) (u *model.User) {
				ctx := context.Background()

				user := &model.CreateUserInput{
					Name:     "test",
					Age:      34,
					Password: "passworddfd",
					Email:    "test@mail.com",
				}
				createdUser, err := repo.Create(ctx, *user)
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
				return createdUser
			},
			act: func(ctx context.Context, t *testing.T, user model.User) (u *model.User, err error) {
				age := 35
				updateUserInput := &model.UpdateUserInput{
					ID:  user.ID,
					Age: &age,
				}
				return repo.Update(ctx, *updateUserInput)
			},
			assert: func(t *testing.T, user *model.User, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 35, user.Age)
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
			// arrange
			createdUser := tt.arrange(t)
			// Act

			updatedUser, err := tt.act(tt.args.ctx, t, *createdUser)

			// Assert
			tt.assert(t, updatedUser, err)

			// teardown

			tt.teardown(t)
		})
	}
}
