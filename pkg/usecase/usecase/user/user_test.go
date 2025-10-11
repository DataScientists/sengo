package usecase_test

import (
	"context"
	"errors"
	"sheng-go-backend/ent/schema/ulid"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/usecase/repository/mocks"
	usecase "sheng-go-backend/pkg/usecase/usecase/user"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupMockUser(t *testing.T) (*mocks.MockUser, func()) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUser(ctrl)
	teardown := func() {
		// Finish will assert that all the expected calls were made.
		ctrl.Finish()
	}
	return mockRepo, teardown
}

// Helper functions to get pointers
func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func TestCreateUser(t *testing.T) {
	mockRepo, teardown := setupMockUser(t)

	defer teardown()

	const (
		ID       = "1"
		PASSWORD = "atestpassword"
		NAME     = "test"
		EMAIL    = "test@mail.com"
	)

	tests := []struct {
		name    string
		input   model.CreateUserInput
		arrange func()
		act     func(uc usecase.User, input model.CreateUserInput) (*model.User, error)
		assert  func(t *testing.T, user *model.User, err error)
	}{
		{
			name: "Should create user",
			input: model.CreateUserInput{
				Name:     NAME,
				Password: PASSWORD,
				Email:    EMAIL,
				Age:      34,
			},
			arrange: func() {
				// Reset expectations for this case
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).
					Return(&model.User{ID: "1", Name: NAME, Email: EMAIL}, nil)
			},
			act: func(uc usecase.User, input model.CreateUserInput) (*model.User, error) {
				return uc.Create(context.Background(), input)
			},
			assert: func(t *testing.T, user *model.User, err error) {
				require.NoError(t, err, "expected no error when creating user")
				require.NotNil(t, user, "expected a non-nil user")
				require.Equal(t, NAME, user.Name, "expected user name to be 'test'")
			},
		}, {
			name: "Should not create user when required fields are missing",
			input: model.CreateUserInput{
				Name:     "",
				Password: "passworddsfd",
				Age:      0,
			},
			arrange: func() {
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
			},
			act: func(uc usecase.User, input model.CreateUserInput) (*model.User, error) {
				return uc.Create(context.Background(), input)
			},
			assert: func(t *testing.T, user *model.User, err error) {
				require.Error(t, err, "expected error when creating user")
				require.Nil(t, user, "expected a nil user")
			},
		},
	}

	// Run the test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: set expectations for this test case.
			tt.arrange()

			// Create the use case with the shared mock repository.
			uc := usecase.NewUserUseCase(mockRepo)

			// Act: call the Create method.
			user, err := tt.act(uc, tt.input)
			// Assert: validate the result.
			tt.assert(t, user, err)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name  string
		input model.UpdateUserInput
		// arrange sets up repository expectations for this test.
		arrange func(ctrl *gomock.Controller, mockRepo *mocks.MockUser)
		act     func(uc usecase.User, input model.UpdateUserInput) (*model.User, error)
		assert  func(t *testing.T, user *model.User, err error)
	}{
		{
			name: "Valid update",
			input: model.UpdateUserInput{
				// For a valid update, the ID must be provided.
				ID:   ulid.MustNew(""),
				Name: strPtr("Updated Name"),
				Age:  intPtr(30),
				// Other fields are optional; here we leave them nil.
			},
			arrange: func(ctrl *gomock.Controller, mockRepo *mocks.MockUser) {
				// Expect that repository.Update is called and returns an updated user.
				mockRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, in model.UpdateUserInput) (*model.User, error) {
						// You can also verify the input if needed.
						// For this example, we return a dummy updated user.
						return &model.User{ID: in.ID, Name: *in.Name}, nil
					})
			},
			act: func(uc usecase.User, input model.UpdateUserInput) (*model.User, error) {
				return uc.Update(context.Background(), input)
			},
			assert: func(t *testing.T, user *model.User, err error) {
				require.NoError(t, err, "expected no error for valid update")
				require.NotNil(t, user, "expected a non-nil user")
				require.Equal(t, "Updated Name", user.Name, "expected the name to be updated")
			},
		},
		{
			name: "Invalid update: missing ID",
			input: model.UpdateUserInput{
				// Missing ID (empty string) should fail validation.
				ID:   "", // ulid.ID is a string alias
				Name: strPtr("Updated Name"),
				Age:  intPtr(30),
			},
			arrange: func(ctrl *gomock.Controller, mockRepo *mocks.MockUser) {
				// Expect that Update is never called because validation should fail.
				mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
			},
			act: func(uc usecase.User, input model.UpdateUserInput) (*model.User, error) {
				return uc.Update(context.Background(), input)
			},
			assert: func(t *testing.T, user *model.User, err error) {
				require.Error(t, err, "expected error when ID is missing")
				require.Nil(t, user, "expected no user to be returned when update fails")
			},
		},
		// You can add more test cases to cover other validation rules.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh gomock controller and mock repository for each sub-test.
			ctrl := gomock.NewController(t)
			mockRepo := mocks.NewMockUser(ctrl)

			// Arrange: set up repository expectations.
			tc.arrange(ctrl, mockRepo)

			// Create the use case instance with the mock repository.
			uc := usecase.NewUserUseCase(mockRepo)

			// Act: call the Update method.
			user, err := tc.act(uc, tc.input)

			// Assert: verify the results.
			tc.assert(t, user, err)

			ctrl.Finish()
		})
	}
}

func TestUserUseCase_Get(t *testing.T) {
	tests := []struct {
		name string
		id   *ulid.ID
		// arrange sets up expectations on the mock repository (if needed)
		arrange func(ctrl *gomock.Controller, mockRepo *mocks.MockUser, id *ulid.ID)
		// act calls the use case's Get method.
		act func(uc usecase.User, id *ulid.ID) (*model.User, error)
		// assert validates the outcome.
		assert func(t *testing.T, user *model.User, err error)
	}{
		{
			name: "nil id returns error",
			id:   nil,
			arrange: func(ctrl *gomock.Controller, mockRepo *mocks.MockUser, id *ulid.ID) {
				// No call should be made on the repository.
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Times(0)
			},
			act: func(uc usecase.User, id *ulid.ID) (*model.User, error) {
				return uc.Get(context.Background(), id)
			},
			assert: func(t *testing.T, user *model.User, err error) {
				require.Error(t, err)
				require.Nil(t, user)
				require.Equal(t, "ID is missing", err.Error())
			},
		},
		{
			name: "empty id returns error",
			id: func() *ulid.ID {
				var empty ulid.ID = ""
				return &empty
			}(),
			arrange: func(ctrl *gomock.Controller, mockRepo *mocks.MockUser, id *ulid.ID) {
				// No call should be made since validation fails before repository call.
			},
			act: func(uc usecase.User, id *ulid.ID) (*model.User, error) {
				return uc.Get(context.Background(), id)
			},
			assert: func(t *testing.T, user *model.User, err error) {
				require.Error(t, err)
				require.Nil(t, user)
				require.Equal(t, "ID is missing", err.Error())
			},
		},
		{
			name: "valid id returns user",
			id: func() *ulid.ID {
				// Create a valid id using your helper (e.g., MustNew)
				valid := ulid.MustNew("user-")
				return &valid
			}(),
			arrange: func(ctrl *gomock.Controller, mockRepo *mocks.MockUser, id *ulid.ID) {
				// Expect that repository.Get is called with the given id
				mockRepo.EXPECT().
					Get(gomock.Any(), id).
					Return(&model.User{ID: *id, Name: "Test User"}, nil)
			},
			act: func(uc usecase.User, id *ulid.ID) (*model.User, error) {
				return uc.Get(context.Background(), id)
			},
			assert: func(t *testing.T, user *model.User, err error) {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.Equal(t, "Test User", user.Name)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new gomock controller and fresh mock repository for each test case.
			ctrl := gomock.NewController(t)
			mockRepo := mocks.NewMockUser(ctrl)

			// Arrange: set expectations if needed.
			if tc.arrange != nil {
				tc.arrange(ctrl, mockRepo, tc.id)
			}

			// Create the use case instance with the mock repository.
			uc := usecase.NewUserUseCase(mockRepo)

			// Act: call the Get method.
			user, err := tc.act(uc, tc.id)

			// Assert: validate the outcome.
			tc.assert(t, user, err)

			ctrl.Finish()
		})
	}
}

func TestUserUseCase_List(t *testing.T) {
	tests := []struct {
		name   string
		after  *model.Cursor
		first  *int
		before *model.Cursor
		last   *int
		where  *model.UserWhereInput
		// arrange sets up expectations on the mock repository for this test case.
		arrange func(ctrl *gomock.Controller, mockRepo *mocks.MockUser, after *model.Cursor, first *int, before *model.Cursor, last *int, where *model.UserWhereInput)
		// act calls the List method.
		act func(uc usecase.User, after *model.Cursor, first *int, before *model.Cursor, last *int, where *model.UserWhereInput) (*model.UserConnection, error)
		// assert checks the result.
		assert func(t *testing.T, conn *model.UserConnection, err error)
	}{
		{
			name:   "List returns connection successfully",
			after:  nil,
			first:  intPtr(10),
			before: nil,
			last:   nil,
			where:  nil,
			arrange: func(ctrl *gomock.Controller, mockRepo *mocks.MockUser, after *model.Cursor, first *int, before *model.Cursor, last *int, where *model.UserWhereInput) {
				// Prepare an expected connection. (Fill in fields as necessary.)
				expectedConn := &model.UserConnection{
					// For example, if UserConnection has an Edges field:
					// Edges: []*model.User{{ID: "1", Name: "Alice"}, {ID: "2", Name: "Bob"}},
				}
				mockRepo.EXPECT().
					List(gomock.Any(), after, first, before, last, where).
					Return(expectedConn, nil)
			},
			act: func(uc usecase.User, after *model.Cursor, first *int, before *model.Cursor, last *int, where *model.UserWhereInput) (*model.UserConnection, error) {
				return uc.List(context.Background(), after, first, before, last, where)
			},
			assert: func(t *testing.T, conn *model.UserConnection, err error) {
				require.NoError(t, err, "expected no error")
				require.NotNil(t, conn, "expected a non-nil connection")
			},
		},
		{
			name:   "List returns error from repository",
			after:  nil,
			first:  intPtr(5),
			before: nil,
			last:   nil,
			where:  nil,
			arrange: func(ctrl *gomock.Controller, mockRepo *mocks.MockUser, after *model.Cursor, first *int, before *model.Cursor, last *int, where *model.UserWhereInput) {
				mockRepo.EXPECT().
					List(gomock.Any(), after, first, before, last, where).
					Return(nil, errors.New("repository error"))
			},
			act: func(uc usecase.User, after *model.Cursor, first *int, before *model.Cursor, last *int, where *model.UserWhereInput) (*model.UserConnection, error) {
				return uc.List(context.Background(), after, first, before, last, where)
			},
			assert: func(t *testing.T, conn *model.UserConnection, err error) {
				require.Error(t, err, "expected an error")
				require.Nil(t, conn, "expected connection to be nil")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh gomock controller and mock repository for each sub-test.
			ctrl := gomock.NewController(t)
			mockRepo := mocks.NewMockUser(ctrl)

			// Arrange: set expectations for this test case.
			tc.arrange(ctrl, mockRepo, tc.after, tc.first, tc.before, tc.last, tc.where)

			// Create the use case with the fresh mock repository.
			uc := usecase.NewUserUseCase(mockRepo)

			// Act: call the List method.
			conn, err := tc.act(uc, tc.after, tc.first, tc.before, tc.last, tc.where)

			// Assert: check the outcome.
			tc.assert(t, conn, err)

			ctrl.Finish()
		})
	}
}
