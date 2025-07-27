package postgres

import (
	"context"
	"database/sql"
	"mizu/internal/db/params"
	"mizu/internal/db/store"

	"github.com/lib/pq"
)

// Postgres implementation of UserStore interface
type ProjectStorePostgres struct {
	db *sql.DB
}

// Creates a new instance of a ProjectStorePostgres
//
// Parameters:
//   - db: *sql.DB
//
// Returns:
//   - *ProjectStorePostgres
func NewProjectStore(db *sql.DB) *ProjectStorePostgres {
	return &ProjectStorePostgres{db: db}
}

func (q *ProjectStorePostgres) CreateOne(ctx context.Context, arg *params.ProjectCreateParams) (int64, error) {

	query := `
	INSERT INTO projects(name, status) 
	VALUE($1, (SELECT id FROM project_statuses WHERE name = $2)) RETURNING id
	`

	var id int64 = 0
	err := q.db.QueryRowContext(
		ctx,
		query,
		arg.Name,
		arg.Status,
	).Scan(&id)

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "unique_violation" {
				return 0, store.ErrUniqueViolation
			}
		}
		return id, store.ErrInsertFailed
	}

	return id, nil
}

// Implementing CreateTask to comply with the ProjectStore interface
func (q *ProjectStorePostgres) CreateTask(ctx context.Context, arg *params.TaskCreateParams) (int64, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, store.ErrInsertFailed
	}

	defer tx.Rollback()

	insertTask := `
	INSERT INTO tasks(project_id, priority, status, name, description, estimated_time_minutes) 
	VALUES($1, (SELECT id FROM task_priorities WHERE name = $2), 
	(SELECT id FROM task_statuses WHERE name = $3), $4, $5, $6) RETURNING id
	`

	var taskId int64 = 0
	err = tx.QueryRowContext(ctx, insertTask,
		arg.ProjectId,
		arg.Priority,
		arg.Status,
		arg.Name,
		arg.Description,
		arg.EstimatedTimeMinutes,
	).Scan(&taskId)

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "unique_violation" {
				return 0, store.ErrUniqueViolation
			}
			if err.Code.Name() == "foreign_key_violation" {
				return 0, store.ErrForeignKeyViolation
			}

			if err.Code.Name() == "not_null_violation" {
				return 0, store.ErrNotNullViolation
			}
		}

		return 0, store.ErrInsertFailed
	}

	insertTaskAssignee := `INSERT INTO task_assignees(task_id, user_id) VALUES($1,$2)`
	for _, assignee := range arg.Assignees {
		if _, err = tx.ExecContext(ctx, insertTaskAssignee, taskId, assignee); err != nil {
			return 0, store.ErrInsertFailed
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, store.ErrInsertFailed
	}

	return taskId, nil

}
