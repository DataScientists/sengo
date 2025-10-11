package registry

import (
	"sheng-go-backend/pkg/adapter/controller"
	userrepository "sheng-go-backend/pkg/adapter/repository/userrepository"
	usecase "sheng-go-backend/pkg/usecase/usecase/user"
)

func (r *registry) NewUserController() controller.User {
	repo := userrepository.NewUserRepository(r.client)
	u := usecase.NewUserUseCase(repo)

	return controller.NewUserController(u)
}
