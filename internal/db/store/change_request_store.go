package store

import (
	"context"
	"mizu/internal/db/params"
)

// ChangeRequestStore abstracts all persistence operations for change requests.
// Lower-level errors should be wrapped in domain-specific error types
// defined in the store package.
// Implementations must be safe for concurrent use.
type ChangeRequestStore interface {
	CreateOne(ctx context.Context, userId int64, arg *params.ChangeRequestCreate) (int64, error)
}
