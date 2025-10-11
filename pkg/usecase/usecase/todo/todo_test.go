package usecase_test

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/usecase/repository/mocks"
	usecase "sheng-go-backend/pkg/usecase/usecase/todo"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupMockTodo(t *testing.T) (*mocks.MockTodo, func()) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockTodo(ctrl)
	teardown := func() {
		// Finish will assert that all the expected calls were made.
		ctrl.Finish()
	}
	return mockRepo, teardown
}

// Helper functions to get pointers
func strPtr(s string) *string                        { return &s }
func intPtr(i int) *int                              { return &i }
func statusPtr(s model.TodoStatus) *model.TodoStatus { return &s }

func TestCreateTodo(t *testing.T) {
	mockRepo, teardown := setupMockTodo(t)

	defer teardown()

	tests := []struct {
		name    string
		input   model.CreateTodoInput
		arrange func()
		act     func(uc usecase.Todo, input model.CreateTodoInput) (*model.Todo, error)
		assert  func(t *testing.T, user *model.Todo, err error)
	}{
		{
			name: "Should create user",
			input: model.CreateTodoInput{
				Name:     strPtr("Test Todo"),
				Priority: intPtr(1),
				Status:   statusPtr("COMPLETED"),
			},
			arrange: func() {
				// Reset expectations for this case
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).
					Return(&model.Todo{Name: "Test Todo", Priority: 1, Status: "IN_PROGRESS"}, nil)
			},
			act: func(uc usecase.Todo, input model.CreateTodoInput) (*model.Todo, error) {
				return uc.Create(context.Background(), input)
			},
			assert: func(t *testing.T, todo *model.Todo, err error) {
				require.NoError(t, err, "expected no error when creating todo")
				require.NotNil(t, todo, "expected a non-nil todo")
				require.Equal(t, "Test Todo", todo.Name, "expected todo name to be 'Test Todo'")
			},
		},
	}

	// Run the test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: set expectations for this test case.
			tt.arrange()

			// Create the use case with the shared mock repository.
			uc := usecase.NewTodoUseCase(mockRepo)

			// Act: call the Create method.
			user, err := tt.act(uc, tt.input)
			// Assert: validate the result.
			tt.assert(t, user, err)
		})
	}
}
