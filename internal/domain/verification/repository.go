package verification

import (
	"context"
	"mizu/internal/domain/iam"
)

type Repository interface {
	Add(ctx context.Context, v *Verification) error

	Get(ctx context.Context, id VerificationID) (*Verification, error)

	Remove(ctx context.Context, v *Verification) error
	RemoveByUserID(ctx context.Context, uID iam.UserID) error
}
