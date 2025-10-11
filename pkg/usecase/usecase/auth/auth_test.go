package usecase_test

import (
	"context"
	"errors"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/usecase/repository/mocks"
	usecase "sheng-go-backend/pkg/usecase/usecase/auth"
	"sheng-go-backend/pkg/util/auth"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupMockAuth(t *testing.T) (*mocks.MockAuth, func()) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockAuth(ctrl)
	teardown := func() {
		// Finish will assert that all the expected calls were made.
		ctrl.Finish()
	}
	return mockRepo, teardown
}

func TestLogin(t *testing.T) {
	const (
		ID       = "1"
		PASSWORD = "atestpassword"
		NAME     = "test"
		EMAIL    = "test@mail.com"
	)

	mockRepo, teardown := setupMockAuth(t)

	defer teardown()

	tests := []struct {
		name    string
		input   model.LoginInput
		arrange func(t *testing.T)
		act     func(uc usecase.Auth, input model.LoginInput) (*model.AuthPayload, error)
		assert  func(t *testing.T, authPayload *model.AuthPayload, err error)
	}{
		{
			name: "Should return valid auth payload",
			input: model.LoginInput{
				Email:    EMAIL,
				Password: PASSWORD,
			},
			arrange: func(t *testing.T) {
				hashedPasswod, err := auth.HashPassword(PASSWORD)
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), EMAIL).
					Return(&model.User{
						ID:       ID,
						Name:     NAME,
						Email:    EMAIL,
						Age:      34,
						Password: hashedPasswod,
					}, nil)
			},
			act: func(uc usecase.Auth, input model.LoginInput) (*model.AuthPayload, error) {
				return uc.Login(context.Background(), input)
			},
			assert: func(t *testing.T, authPayload *model.AuthPayload, err error) {
				assert.Nil(t, err, "Error  should be nil")
				assert.NotNil(t, authPayload, "Authpayload should not be nil")
				assert.NotNil(t, authPayload.AccessToken, "AcessToken should not be nil")
				assert.NotNil(t, authPayload.RefreshToken, "RefreshToken should not be nil")
				assert.NotNil(t, authPayload.User, "User should not be nil")
			},
		},
		{
			name: "Should return email is required when  email is missing",
			input: model.LoginInput{
				Email:    "",
				Password: "",
			},
			arrange: func(t *testing.T) {
				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), EMAIL).Times(0)
			},
			act: func(uc usecase.Auth, input model.LoginInput) (*model.AuthPayload, error) {
				return uc.Login(context.Background(), input)
			},
			assert: func(t *testing.T, authPayload *model.AuthPayload, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, "Email is required", err.Error())
			},
		},
		{
			name: "Should return invalid credentails when record is missing",
			input: model.LoginInput{
				Email:    EMAIL,
				Password: PASSWORD,
			},
			arrange: func(t *testing.T) {
				mockRepo.EXPECT().
					GetUserByEmail(gomock.Any(), EMAIL).
					Return(nil, errors.New("User Not Found"))
			},
			act: func(uc usecase.Auth, input model.LoginInput) (*model.AuthPayload, error) {
				return uc.Login(context.Background(), input)
			},
			assert: func(t *testing.T, authPayload *model.AuthPayload, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, "Invalid Credentials", err.Error())
			},
		},
		{
			name: "Should return invalid credentials when password is incorrect",
			input: model.LoginInput{
				Email:    EMAIL,
				Password: "wrongpassword",
			},
			arrange: func(t *testing.T) {
				hashedPassword, err := auth.HashPassword(PASSWORD)
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
				mockRepo.EXPECT().
					GetUserByEmail(gomock.Any(), EMAIL).
					// Ensure EMAIL is passed correctly
					Return(&model.User{
						ID:       ID,
						Name:     NAME,
						Email:    EMAIL,
						Age:      34,
						Password: hashedPassword,
					}, nil)
			},
			act: func(uc usecase.Auth, input model.LoginInput) (*model.AuthPayload, error) {
				return uc.Login(context.Background(), input)
			},
			assert: func(t *testing.T, authPayload *model.AuthPayload, err error) {
				assert.NotNil(t, err, "Error should not be nil")
				assert.Equal(t, "Invalid Credentials", err.Error()) // Corrected typo
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authUseCase := usecase.NewAuthUseCase(mockRepo)
			tt.arrange(t)
			authPayload, err := tt.act(authUseCase, tt.input)
			tt.assert(t, authPayload, err)
		})
	}
}
