package todorepository

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
)

func (r *todoRepository) Create(
	ctx context.Context,
	input model.CreateTodoInput,
) (*model.Todo, error) {
	todo, err := r.client.Todo.Create().SetInput(input).Save(ctx)
	if err != nil {
		return nil, model.NewDBError(err)
	}
	return todo, nil
}
