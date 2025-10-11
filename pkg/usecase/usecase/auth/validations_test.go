package usecase_test

import (
	"sheng-go-backend/pkg/entity/model"
	usecase "sheng-go-backend/pkg/usecase/usecase/auth"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLoginInput(t *testing.T) {
	tests := []struct {
		name    string
		arrange func() model.LoginInput
		act     func(input model.LoginInput) error
		assert  func(t *testing.T, err error)
	}{
		{
			name: "Should pass input is valid",
			arrange: func() model.LoginInput {
				return model.LoginInput{
					Email:    "bhuwan@beyul.com",
					Password: "Password12345",
				}
			},
			act: func(input model.LoginInput) error {
				return usecase.ValidateLoginInput(input)
			},
			assert: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "Missing Email",
			arrange: func() model.LoginInput {
				return model.LoginInput{
					Email:    "",
					Password: "Password12345",
				}
			},
			act: func(input model.LoginInput) error {
				return usecase.ValidateLoginInput(input)
			},
			assert: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, "Email is required", err.Error())
			},
		},
		{
			name: "Missing Password",
			arrange: func() model.LoginInput {
				return model.LoginInput{
					Email:    "bhuwan@mail.com",
					Password: "",
				}
			},
			act: func(input model.LoginInput) error {
				return usecase.ValidateLoginInput(input)
			},
			assert: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, "Password is required", err.Error())
			},
		},
		{
			name: "Password length is less than 8",
			arrange: func() model.LoginInput {
				return model.LoginInput{
					Email:    "bhuwan@mail.com",
					Password: "dsafds",
				}
			},
			act: func(input model.LoginInput) error {
				return usecase.ValidateLoginInput(input)
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
