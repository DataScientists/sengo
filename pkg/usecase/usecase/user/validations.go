package usecase

import (
	"errors"
	"sheng-go-backend/pkg/entity/model"
)

// validateCreateUserInput checks the required fields of CreateUserInput.
func ValidateCreateUserInput(input model.CreateUserInput) error {
	if input.Email == "" {
		return errors.New("Email is required")
	}

	if input.Name == "" {
		return errors.New("Name is required")
	}
	if input.Age <= 0 {
		return errors.New("Age must be a positive number")
	}

	if len(input.Password) < 8 {
		return errors.New("Password must be at least 8 characters long")
	}
	return nil
}

// validateUpdateUserInput checks that the UpdateUserInput satisfies required business rules.
// It returns an error if any validation fails.
func ValidateUpdateUserInput(input model.UpdateUserInput) error {
	// Check that the ID is provided (assuming the zero value of ulid.ID indicates an empty ID).
	if input.ID == "" {
		return errors.New("ID is required")
	}

	// If Name is provided, it must not be empty.
	if input.Name != nil && *input.Name == "" {
		return errors.New("if name is provided, it cannot be empty")
	}

	// If Age is provided, it must be positive.
	if input.Age != nil && *input.Age <= 0 {
		return errors.New("if age is provided, it must be a positive number")
	}

	return nil
}
