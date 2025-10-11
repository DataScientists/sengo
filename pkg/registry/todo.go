package registry

import (
	"sheng-go-backend/pkg/adapter/controller"
	todorepository "sheng-go-backend/pkg/adapter/repository/todorepository"
	usecase "sheng-go-backend/pkg/usecase/usecase/todo"
)

func (r *registry) NewTodoController() controller.Todo {
	repo := todorepository.NewTodoRepository(r.client)
	u := usecase.NewTodoUseCase(repo)

	return controller.NewTodoController(u)
}
