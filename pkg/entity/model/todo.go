package model

import (
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/todo"
)

// Todo is the model entity for the Todo schema.
type Todo = ent.Todo

// CreateTodoInput represents a mutation input for creating todos.
type CreateTodoInput = ent.CreateTodoInput

// UpdateTodoInput represents a mutation input for updating todos.
type UpdateTodoInput = ent.UpdateTodoInput

type TodoWhereInput = ent.TodoWhereInput

type TodoConnection = ent.TodoConnection

type TodoStatus = todo.Status
