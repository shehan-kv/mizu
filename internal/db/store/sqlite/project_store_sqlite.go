package sqlite

import (
	"context"
	"database/sql"
	agg "mizu/internal/db/models/aggregates"
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

func (q *ProjectStoreSqlite) CreateOne(ctx context.Context, arg *params.ProjectCreate) (int64, error) {

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
func (q *ProjectStoreSqlite) CreateTask(ctx context.Context, arg *params.TaskCreate) (int64, error) {

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

// Implementing GetWithStats defined in ProjectStore interface
func (q *ProjectStoreSqlite) GetWithStats(
	ctx context.Context,
	arg *params.ProjectsSearch) (*agg.WithCount[agg.ProjectWithStats], error) {

	query := `
	SELECT p.id, p.name, p.created_at, ps.name AS status,
  
  	COALESCE(t.total_tasks, 0) AS total_tasks,
  	COALESCE(t.completed_tasks, 0) AS tasks_completed,
  
  	COALESCE(i.total_invoices, 0) AS total_invoices,
  	COALESCE(i.paid_invoices, 0) AS invoices_paid,
   	COALESCE(i.total_quotes, 0) AS total_quotes

	FROM projects p
	JOIN project_statuses ps ON p.status = ps.id

	LEFT JOIN (
  	SELECT 
    	t.project_id,
    	COUNT(*) AS total_tasks,
    	COUNT(CASE WHEN ts.name = 'completed' THEN 1 END) AS completed_tasks
  	FROM tasks t
  	JOIN task_statuses ts ON t.status = ts.id
  	GROUP BY t.project_id
	) t ON p.id = t.project_id

	LEFT JOIN (
  	SELECT
    	i.project_id,
    	COUNT(CASE WHEN i.is_invoice = TRUE THEN 1 END) AS total_invoices,
		COUNT(CASE WHEN i.is_invoice = FALSE THEN 1 END) AS total_quotes,
    	COUNT(CASE WHEN ins.name = 'paid' AND i.is_invoice = TRUE THEN 1 END) AS paid_invoices
  	FROM invoices i
  	JOIN invoice_statuses ins ON i.status = ins.id
  	GROUP BY i.project_id
	) i ON p.id = i.project_id

	LEFT JOIN project_users pu ON p.id = pu.project_id
	WHERE pu.user_id = ?
	`

	projectCount := `
	SELECT COUNT(p.id) FROM projects p 
	JOIN project_users pu ON p.id = pu.project_id
	JOIN project_statuses ps ON ps.id = p.status 
	WHERE pu.user_id = ?
	`

	queryArgs := []any{arg.UserId}
	countArgs := []any{arg.UserId}

	if !(arg.Status == "") {
		query += " AND ps.name = ?"
		projectCount += " AND ps.name = ?"
		queryArgs = append(queryArgs, arg.Status)
		countArgs = append(countArgs, arg.Status)
	}

	if !(arg.Keyword == "") {
		query += " AND p.name LIKE ?"
		projectCount += " AND p.name LIKE ?"
		queryArgs = append(queryArgs, "%"+arg.Keyword+"%")
		countArgs = append(countArgs, "%"+arg.Keyword+"%")
	}

	query += " LIMIT ? OFFSET ?"
	queryArgs = append(queryArgs, arg.Limit, arg.Offset)

	var totalProjects int64 = 10
	err := q.db.QueryRowContext(ctx, projectCount, countArgs...).Scan(&totalProjects)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	rows, err := q.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	result := agg.WithCount[agg.ProjectWithStats]{
		Total: totalProjects,
		Items: make([]agg.ProjectWithStats, 0),
	}

	for rows.Next() {
		var project agg.ProjectWithStats

		err := rows.Scan(
			&project.Id,
			&project.Name,
			&project.CreatedAt,
			&project.Status,
			&project.TotalTasks,
			&project.TasksCompleted,
			&project.TotalInvoices,
			&project.InvoicesPaid,
			&project.TotalQuotes,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		result.Items = append(result.Items, project)
	}

	return &result, nil
}
