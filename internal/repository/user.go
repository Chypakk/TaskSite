package repository

import (
	"context"
	"tasksite/internal/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, username, password string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserById(ctx context.Context, id int) (*model.User, error)
}