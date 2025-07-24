package store

import (
	"context"
	"mizu/internal/db/params"
)

type ProjectStore interface {

	// Creates a project, assigns users,
	// creates a channel and assigns users to the channel.
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - arg: pointer to ProjectCreateParams
	//
	// Returns:
	//   - int64: id of new project
	//   - errdefs.ErrDbUniqueViolation if name already exists
	//   - errdefs.ErrDbInsertFailed if create fails
	CreateOne(ctx context.Context, arg *params.ProjectCreateParams) (int64, error)
}
