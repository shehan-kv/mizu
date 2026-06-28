package iam

import "context"

type RecoveryRepository interface {
	Add(ctx context.Context, r *Recovery) error

	GetByToken(ctx context.Context, t RecoveryToken) (*Recovery, error)
	GetByUser(ctx context.Context, uID UserID) (*Recovery, error)

	Save(ctx context.Context, r *Recovery) error

	Remove(ctx context.Context, r *Recovery) error
	RemoveByUser(ctx context.Context, uID UserID) error
}
