package usecase_test

import (
	"sheng-go-backend/pkg/entity/model"
	usecase "sheng-go-backend/pkg/usecase/usecase/user"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCreateUserInput(t *testing.T) {
	tests := []struct {
		name    string
		arrange func() model.CreateUserInput
		act     func(input model.CreateUserInput) error
		assert  func(t *testing.T, err error)
	}{
		{
			name: "Should pass input is valid",
			arrange: func() model.CreateUserInput {
				return model.CreateUserInput{
					Email:    "bhuwan@beyul.com",
					Name:     "Bhuwan",
					Age:      35,
					Password: "Password12345",
				}
			},
			act: func(input model.CreateUserInput) error {
				return usecase.ValidateCreateUserInput(input)
			},
			assert: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "Missing Name",
			arrange: func() model.CreateUserInput {
				return model.CreateUserInput{
					Email:    "bhuwan@beyul.com",
					Name:     "",
					Age:      35,
					Password: "Password12345",
				}
			},
			act: func(input model.CreateUserInput) error {
				return usecase.ValidateCreateUserInput(input)
			},
			assert: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, "Name is required", err.Error())
			},
		},
		{
			name: "Invalid Age (zero)",
			arrange: func() model.CreateUserInput {
				return model.CreateUserInput{
					Email:    "bhuwan@beyul.com",
					Name:     "Bhuwan",
					Age:      0,
					Password: "Password12345",
				}
			},
			act: func(input model.CreateUserInput) error {
				return usecase.ValidateCreateUserInput(input)
			},
			assert: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, "Age must be a positive number", err.Error())
			},
		},
		{
			name: "Invalid Age (negetive)",
			arrange: func() model.CreateUserInput {
				return model.CreateUserInput{
					Email:    "bhuwan@beyul.com",
					Name:     "Bhuwan",
					Age:      -2,
					Password: "Password12345",
				}
			},
			act: func(input model.CreateUserInput) error {
				return usecase.ValidateCreateUserInput(input)
			},
			assert: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, "Age must be a positive number", err.Error())
			},
		},
		{
			name: "Password Too Short",
			arrange: func() model.CreateUserInput {
				return model.CreateUserInput{
					Email:    "bhuwan@beyul.com",
					Name:     "Bhuwan",
					Age:      2,
					Password: "Pass345",
				}
			},
			act: func(input model.CreateUserInput) error {
				return usecase.ValidateCreateUserInput(input)
			},
			assert: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, "Password must be at least 8 characters long", err.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.arrange()
			err := tt.act(input)
			tt.assert(t, err)
		})
	}
}
