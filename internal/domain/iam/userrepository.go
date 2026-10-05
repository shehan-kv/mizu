package iam

import (
	"context"
	"mizu/internal/domain/common"
)

type UserRepository interface {
	Add(ctx context.Context, user *User) error

	Exists(ctx context.Context, id UserID) (bool, error)
	ExistsAll(ctx context.Context, ids []UserID) (bool, error)

	GetByID(ctx context.Context, id UserID) (*User, error)
	GetByEmail(ctx context.Context, e Email) (*User, error)
	List(ctx context.Context, filter UserFilter, page common.Page) ([]*User, error)
	ListByIDs(ctx context.Context, ids []UserID, f UserFilter) ([]*User, error)

	IsAnyAdministrator(ctx context.Context, ids []UserID) (bool, error)
	HasAdministrator(ctx context.Context) (bool, error)

	Count(ctx context.Context, filter UserFilter) (int, error)

	Save(ctx context.Context, user *User) error

	Remove(ctx context.Context, user *User) error
}
