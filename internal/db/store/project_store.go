package store

import (
	"context"
	agg "mizu/internal/db/models/aggregates"
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
	CreateOne(ctx context.Context, arg *params.ProjectCreate) (int64, error)

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
	CreateTask(ctx context.Context, arg *params.TaskCreate) (int64, error)

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
	GetWithStats(ctx context.Context, arg *params.ProjectsSearch) (*agg.WithCount[agg.ProjectWithStats], error)

	GetById(ctx context.Context, projectId int64) (*agg.Project, error)

	GetTasksByProjectId(ctx context.Context, projectId int64, arg *params.TaskSearch) ([]agg.Task, error)

	CountTasksByProjectId(ctx context.Context, projectId int64, arg *params.TaskSearch) (int64, error)

	GetMembersByProjectId(ctx context.Context, projectId int64) ([]agg.ProjectUser, error)

	GetTaskMetricsByProjectId(
		ctx context.Context,
		projectId int64,
		status params.TaskStatus) ([]agg.TaskMetric, error)
}
