package registry

import (
	"sheng-go-backend/pkg/adapter/controller"
	"sheng-go-backend/pkg/adapter/repository/authrepository"
	usecase "sheng-go-backend/pkg/usecase/usecase/auth"
)

func (r *registry) NewAuthController() controller.Auth {
	repo := authrepository.NewAuthRepository(r.client)
	u := usecase.NewAuthUseCase(repo)

	return controller.NewAuthController(u)
}
