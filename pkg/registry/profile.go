package registry

import (
	"sheng-go-backend/pkg/adapter/controller"
	profilerepository "sheng-go-backend/pkg/adapter/repository/profilerepository"
	usecase "sheng-go-backend/pkg/usecase/usecase/profile"
)

func (r *registry) NewProfileController() controller.Profile {
	repo := profilerepository.NewProfileRepository(r.client)
	u := usecase.NewProfileUseCase(repo)

	return controller.NewProfileController(u)
}
