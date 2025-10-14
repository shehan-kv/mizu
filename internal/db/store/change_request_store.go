package store

import (
	"context"
	agg "mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
)

// ChangeRequestStore abstracts all persistence operations for change requests.
// Lower-level errors should be wrapped in domain-specific error types
// defined in the store package.
// Implementations must be safe for concurrent use.
type ChangeRequestStore interface {
	CreateOne(ctx context.Context, userId int64, arg *params.ChangeRequestCreate) (int64, error)

	CreateEntry(ctx context.Context, userId int64, requestId int64, content string) error

	GetByProjectId(ctx context.Context, projectId int64, arg *params.ChangeRequestSearch) (*agg.WithCount[agg.ChangeRequest], error)

	CloseById(ctx context.Context, requestId int64) error

	GetEntriesByRequestId(ctx context.Context, requestId int64) ([]agg.ChangeRequestEntry, error)

	GetById(ctx context.Context, requestId int64) (*agg.ChangeRequest, error)

	GetByUserId(ctx context.Context, userId int64, arg *params.ChangeRequestSearch) ([]agg.ChangeRequest, error)

	CountByUserId(ctx context.Context, userId int64, arg *params.ChangeRequestSearch) (int64, error)

	CountByProjectId(ctx context.Context, projectId int64, arg *params.ChangeRequestSearch) (int64, error)
}
