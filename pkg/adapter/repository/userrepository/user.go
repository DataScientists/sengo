// This is implement the use case and match to its types
package userrepository

import (
	"sheng-go-backend/ent"
	ur "sheng-go-backend/pkg/usecase/repository"
)

type userRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) ur.User {
	return &userRepository{client: client}
}
