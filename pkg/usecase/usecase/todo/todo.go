package usecase

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/usecase/repository"
)

type todoUseCase struct {
	todoRepository repository.Todo
}

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

// This function creates new todo use case
func NewTodoUseCase(r repository.Todo) Todo {
	return &todoUseCase{todoRepository: r}
}

func (t *todoUseCase) Get(ctx context.Context, where *model.TodoWhereInput) (*model.Todo, error) {
	return t.todoRepository.Get(ctx, where)
}

func (t *todoUseCase) Create(
	ctx context.Context,
	input model.CreateTodoInput,
) (*model.Todo, error) {
	return t.todoRepository.Create(ctx, input)
}

func (t *todoUseCase) Update(
	ctx context.Context,
	input model.UpdateTodoInput,
) (*model.Todo, error) {
	return t.todoRepository.Update(ctx, input)
}

func (t *todoUseCase) List(
	ctx context.Context,
	after *model.Cursor,
	first *int,
	before *model.Cursor,
	last *int, where *model.TodoWhereInput,
) (*model.TodoConnection, error) {
	return t.todoRepository.List(ctx, after, first, before, last, where)
}
