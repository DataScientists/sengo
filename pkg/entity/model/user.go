package model

import "sheng-go-backend/ent"

// User is the model entity for the User schema.
type User = ent.User

// CreateUserInput represents a mutation input for creating users.
type CreateUserInput = ent.CreateUserInput

// UpdateUserInput represents a mutation input for updating users.
type UpdateUserInput = ent.UpdateUserInput

// UserConnection is the connnection containing edges to User.
type UserConnection = ent.UserConnection

// UserWhereInput represents a where input for filtering user queries
type UserWhereInput = ent.UserWhereInput
