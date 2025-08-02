package sqlite

import (
	"context"
	"database/sql"
	"mizu/internal/db/params"
	"mizu/internal/db/store"

	"github.com/mattn/go-sqlite3"
)

// SQLite implementation of ProjectStore interface
type ProjectStoreSqlite struct {
	db *sql.DB
}

// Creates a new instance of a ProjectStoreSqlite
//
// Parameters:
//   - db: *sql.DB
//
// Returns:
//   - *ProjectStoreSqlite
func NewProjectStore(db *sql.DB) *ProjectStoreSqlite {
	return &ProjectStoreSqlite{db: db}
}

func (q *ProjectStoreSqlite) CreateOne(ctx context.Context, arg *params.ProjectCreateParams) (int64, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, store.ErrInsertFailed
	}

	defer tx.Rollback()

	insertProject := `
	INSERT INTO projects(name, status) 
	VALUES(?, (SELECT id FROM project_statuses WHERE name = ?)) RETURNING id
	`

	var projectId int64 = 0
	err = tx.QueryRowContext(ctx, insertProject, arg.Name, arg.Status).Scan(&projectId)

	if err != nil {
		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
				return 0, store.ErrUniqueViolation
			}
		}

		return 0, store.ErrInsertFailed
	}

	var channelId int64 = 0
	insertChannel := `INSERT INTO channels(project_id, name) VALUES(?,?) RETURNING id`
	if err = tx.QueryRowContext(ctx, insertChannel, projectId, arg.Name).Scan(&channelId); err != nil {
		return 0, store.ErrInsertFailed
	}

	insertProjectUser := `INSERT INTO project_users(user_id, project_id) VALUES(?,?)`
	for _, memberId := range arg.Members {
		if _, err = tx.ExecContext(ctx, insertProjectUser, memberId, projectId); err != nil {
			return 0, store.ErrInsertFailed
		}
	}

	insertChannelUser := `INSERT INTO channel_users(channel_id, user_id) VALUES(?,?)`
	for _, memberId := range arg.Members {
		if _, err = tx.ExecContext(ctx, insertChannelUser, channelId, memberId); err != nil {
			return 0, store.ErrInsertFailed
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, store.ErrInsertFailed
	}

	return projectId, nil
}

// Implementing CreateTask defined in ProjectStore interface
func (q *ProjectStoreSqlite) CreateTask(ctx context.Context, arg *params.TaskCreateParams) (int64, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, store.ErrInsertFailed
	}

	defer tx.Rollback()

	insertTask := `
	INSERT INTO tasks(project_id, priority, status, name, description, estimated_time_minutes) 
	VALUES(?, (SELECT id FROM task_priorities WHERE name = ?), 
	(SELECT id FROM task_statuses WHERE name = ?), ?, ?, ?) RETURNING id
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
		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
				return 0, store.ErrUniqueViolation
			}
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return 0, store.ErrForeignKeyViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
				return 0, store.ErrNotNullViolation
			}
		}

		return 0, store.ErrInsertFailed
	}

	insertTaskAssignee := `INSERT INTO task_assignees(task_id, user_id) VALUES(?,?)`
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
