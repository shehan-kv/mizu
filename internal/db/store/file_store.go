package store

import (
	"context"
	agg "mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
)

// FileStore abstracts all persistence operations for files.
// Lower-level errors should be wrapped in domain-specific error types
// defined in the store package.
// Implementations must be safe for concurrent use.
type FileStore interface {
	CreateOne(ctx context.Context, arg params.FileCreate) (int64, error)

	GetByChannelId(
		ctx context.Context,
		channelId int64,
		arg *params.FileSearch) (*agg.WithCount[agg.FileWithUser], error)

	GetByProjectId(
		ctx context.Context,
		projectId int64,
		arg *params.FileSearch) ([]agg.FileWithUser, error)

	CountByProjectId(ctx context.Context, projectId int64, arg *params.FileSearch) (int64, error)
}
