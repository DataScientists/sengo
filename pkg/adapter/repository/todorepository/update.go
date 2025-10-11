package todorepository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/entity/model"
)

func (r *todoRepository) Update(
	ctx context.Context,
	input model.UpdateTodoInput,
) (*model.Todo, error) {
	u, err := r.client.Todo.UpdateOneID(input.ID).SetInput(input).Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, model.NewNotFoundError(err, input.ID)
		}

		return nil, model.NewDBError(err)
	}
	return u, nil
}
