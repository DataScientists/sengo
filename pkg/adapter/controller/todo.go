package controller

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
	usecase "sheng-go-backend/pkg/usecase/usecase/todo"
)

type Todo interface {
	Get(ctx context.Context, where *model.TodoWhereInput) (*model.Todo, error)
	Create(ctx context.Context, input model.CreateTodoInput) (*model.Todo, error)
	Update(ctx context.Context, input model.UpdateTodoInput) (*model.Todo, error)
	List(ctx context.Context,
		after *model.Cursor,
		first *int,
		before *model.Cursor,
		last *int, where *model.TodoWhereInput) (*model.TodoConnection, error)
}

type todoController struct {
	todoUseCase usecase.Todo
}

// Create new todo controller

func NewTodoController(tu usecase.Todo) Todo {
	return &todoController{todoUseCase: tu}
}

func (tc *todoController) Get(
	ctx context.Context,
	where *model.TodoWhereInput,
) (*model.Todo, error) {
	return tc.todoUseCase.Get(ctx, where)
}

func (tc *todoController) Create(
	ctx context.Context,
	input model.CreateTodoInput,
) (*model.Todo, error) {
	return tc.todoUseCase.Create(ctx, input)
}

func (tc *todoController) Update(
	ctx context.Context,
	input model.UpdateTodoInput,
) (*model.Todo, error) {
	return tc.todoUseCase.Update(ctx, input)
}

func (tc *todoController) List(
	ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int, where *model.TodoWhereInput,
) (*model.TodoConnection, error) {
	return tc.todoUseCase.List(ctx, after, first, before, last, where)
}
