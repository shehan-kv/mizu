package postgres

import (
	"context"
	"database/sql"
	agg "mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"strconv"

	"github.com/lib/pq"
)

// Postgres implementation of ProjectStore interface
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

func (q *ProjectStorePostgres) CreateOne(ctx context.Context, arg *params.ProjectCreate) (int64, error) {

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
func (q *ProjectStorePostgres) CreateTask(ctx context.Context, arg *params.TaskCreate) (int64, error) {

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

// Implementing GetWithStats defined in ProjectStore interface
func (q *ProjectStorePostgres) GetWithStats(
	ctx context.Context,
	arg *params.ProjectsSearch) (*agg.ProjectWithStatsList, error) {

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
	WHERE pu.user_id = $1
	`

	projectCount := `
	SELECT COUNT(p.id) FROM projects p JOIN project_users pu ON p.id = pu.project_id
	WHERE pu.user_id = $1
	`

	queryArgs := []any{arg.UserId}
	countArgs := []any{arg.UserId}

	paramCount := 1

	if !(arg.Status == "") {
		paramCount++
		strParamCount := strconv.Itoa(paramCount)
		query += " AND ps.name = $" + strParamCount
		projectCount += " AND ps.name = $" + strParamCount
		queryArgs = append(queryArgs, arg.Status)
		countArgs = append(countArgs, arg.Status)
	}

	if !(arg.Keyword == "") {
		paramCount++
		strParamCount := strconv.Itoa(paramCount)
		query += " AND p.name LIKE $" + strParamCount
		projectCount += " AND p.name LIKE $" + strParamCount
		queryArgs = append(queryArgs, "%"+arg.Keyword+"%")
		countArgs = append(countArgs, "%"+arg.Keyword+"%")
	}

	paramCount++
	strLimitParamCount := strconv.Itoa(paramCount)
	paramCount++
	strOffsetParamCount := strconv.Itoa(paramCount)

	query += " LIMIT $" + strLimitParamCount + " OFFSET $" + strOffsetParamCount
	queryArgs = append(queryArgs, arg.Limit, arg.Offset)

	rows, err := q.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	projects := []agg.ProjectWithStats{}

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

		projects = append(projects, project)
	}

	var totalProjects int64
	err = q.db.QueryRowContext(ctx, projectCount, countArgs...).Scan(&totalProjects)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	return &agg.ProjectWithStatsList{TotalCount: totalProjects, Projects: projects}, nil
}
