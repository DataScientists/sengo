package registry

import (
	"sheng-go-backend/pkg/adapter/controller"
	profileentryrepository "sheng-go-backend/pkg/adapter/repository/profileentryrepository"
	usecase "sheng-go-backend/pkg/usecase/usecase/profileentry"
)

func (r *registry) NewProfileEntryController() controller.ProfileEntry {
	repo := profileentryrepository.NewprofileentryRepository(r.client)
	u := usecase.NewProfileEntryUseCase(repo)

	return controller.NewProfileEntryController(u)
}
