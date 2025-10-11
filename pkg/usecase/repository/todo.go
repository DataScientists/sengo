//go:generate mockgen -source=todo.go -destination=./mocks/todo_repository_mock.go -package=mocks
package repository

import (
	"context"
	"sheng-go-backend/pkg/entity/model"
)

// Todo is an interface of repository

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
