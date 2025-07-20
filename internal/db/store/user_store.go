package store

import (
	"context"
	"mizu/internal/db/models"
	"mizu/internal/db/params"
)

// Defines the behavior required for managing users
type UserStore interface {
	CreateOne(ctx context.Context, arg *params.UserCreateParams) (int64, error)
	GetById(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetPasswordById(ctx context.Context, id int64) (string, error)
	CountAll(ctx context.Context) (int64, error)
	SetPasswordById(ctx context.Context, id int64, password string) error
	DeleteById(ctx context.Context, id int64) error
	UpdateLastLogin(ctx context.Context, id int64) error
}
