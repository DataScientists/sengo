package todorepository

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/entity/model"
)

func (r *todoRepository) Get(
	ctx context.Context,
	where *model.TodoWhereInput,
) (*model.Todo, error) {
	q := r.client.Todo.Query()

	q, err := where.Filter(q)
	if err != nil {
		return nil, model.NewInvalidParamError(nil)
	}

	res, err := q.Only(ctx)
	if err != nil {
		if ent.IsNotSingular(err) {
			return nil, model.NewNotFoundError(err, nil)
		}
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, model.NewDBError(err)
	}

	return res, nil
}

func (r *todoRepository) List(
	ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int,
	where *model.TodoWhereInput,
) (*model.TodoConnection, error) {
	tc, err := r.client.Todo.Query().
		Paginate(ctx, after, first, before, last, ent.WithTodoFilter(where.Filter))
	if err != nil {
		return nil, model.NewDBError(err)
	}

	return tc, nil
}
