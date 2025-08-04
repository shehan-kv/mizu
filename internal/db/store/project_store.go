package store

import (
	"context"
	"mizu/internal/db/models/aggregates"
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
	//   - store.ErrUniqueViolation: if name already exists
	//   - store.ErrInsertFailed: if create fails
	CreateOne(ctx context.Context, arg *params.ProjectCreateParams) (int64, error)

	// Creates a project task, assigns users if needed,
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - arg: pointer to TaskCreateParams
	//
	// Returns:
	//   - int64: id of new project
	//   - store.ErrUniqueViolation: if name already exists in the same project
	//   - store.ErrForeignKeyViolation: if foreign key is invalid
	//	 - store.ErrNotNullViolation: if not-null constrain violated
	//   - store.ErrInsertFailed: if create fails
	CreateTask(ctx context.Context, arg *params.TaskCreateParams) (int64, error)

	// Gets a list of projects with:
	// 	 - number of tasks
	//	 - number of completed tasks
	//	 - number of invoices
	//	 - number of paid invoices
	// 	 - number of quotes
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - arg: pointer to ProjectsSearchParams
	//
	// Returns:
	//	 - *aggregates.ProjectWithStatsList
	//   - store.ErrQueryFailed: if query fails
	GetWithStats(ctx context.Context, arg *params.ProjectsSearchParams) (*aggregates.ProjectWithStatsList, error)
}
