package usecase

import (
	"errors"
	"sheng-go-backend/pkg/entity/model"
)

func ValidateLoginInput(input model.LoginInput) error {
	if input.Email == "" {
		return errors.New("Email is required")
	}
	if input.Password == "" {
		return errors.New("Password is required")
	}
	if len(input.Password) < 8 {
		return errors.New("Password must be at least 8 characters long")
	}
	// Additional validations as needed.
	return nil
}
