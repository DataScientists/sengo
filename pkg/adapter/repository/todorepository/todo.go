package todorepository

import (
	"sheng-go-backend/ent"
	ur "sheng-go-backend/pkg/usecase/repository"
)

type todoRepository struct {
	client *ent.Client
}

func NewTodoRepository(client *ent.Client) ur.Todo {
	return &todoRepository{client}
}
