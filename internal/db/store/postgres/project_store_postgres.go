package postgres

import (
	"context"
	"database/sql"
	agg "mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"strconv"
	"strings"

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
	WHERE pu.user_id = $1
	`

	projectCount := `
	SELECT COUNT(p.id) FROM projects p 
	JOIN project_users pu ON p.id = pu.project_id
	JOIN project_statuses ps ON ps.id = p.status 
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

	var totalProjects int64
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

func (q *ProjectStorePostgres) GetTasksByProjectId(
	ctx context.Context,
	projectId int64,
	arg *params.TaskSearch) ([]agg.Task, error) {

	var query strings.Builder
	query.WriteString(`
	WITH paged_tasks AS (
		SELECT 
			t.id,
			t.project_id,
			t.name,
			t.description,
			t.created_at,
			t.estimated_time_minutes,
			ts.name AS status,
			tp.name AS priority
		FROM tasks t
		JOIN task_statuses ts ON ts.id = t.status 
		JOIN task_priorities tp ON tp.id = t.priority
		WHERE t.project_id = $1
	`)

	queryArgs := []any{projectId}

	paramCount := 1

	if len(arg.Keyword) != 0 {
		paramCount++
		query.WriteString(" AND t.name LIKE $")
		query.WriteString(strconv.Itoa(paramCount))
		queryArgs = append(queryArgs, "%"+arg.Keyword+"%")
	}

	if len(arg.Status) != 0 {
		paramCount++
		query.WriteString(" AND ts.name = $")
		query.WriteString(strconv.Itoa(paramCount))
		queryArgs = append(queryArgs, arg.Status)
	}

	if len(arg.Priority) != 0 {
		paramCount++
		query.WriteString(" AND tp.name = $")
		query.WriteString(strconv.Itoa(paramCount))
		queryArgs = append(queryArgs, arg.Priority)
	}

	query.WriteString(" ORDER BY t.created_at DESC")

	paramCount++
	query.WriteString(" LIMIT $")
	query.WriteString(strconv.Itoa(paramCount))

	paramCount++
	query.WriteString(" OFFSET $")
	queryArgs = append(queryArgs, arg.Limit, arg.Offset)
	query.WriteString(strconv.Itoa(paramCount))
	query.WriteString(" )") // Closing 'WITH paged_tasks AS'

	query.WriteString(`
	SELECT 
		t.id,
		t.project_id,
		t.name,
		t.description,
		t.created_at,
		t.estimated_time_minutes,
		t.status,
		t.priority,
		u.id,
		u.first_name,
		u.last_name,
		u.image,
		u.title
	FROM paged_tasks t
	LEFT JOIN task_assignees ta ON ta.task_id = t.id
	LEFT JOIN users u ON u.id = ta.user_id
	`)

	rows, err := q.db.QueryContext(ctx, query.String(), queryArgs...)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	tasks := make([]agg.Task, 0)

	for rows.Next() {
		var row agg.Task

		err := rows.Scan(
			&row.Id,
			&row.ProjectId,
			&row.Name,
			&row.Description,
			&row.CreatedAt,
			&row.EstTimeMinutes,
			&row.Status,
			&row.Priority,
			&row.UserId,
			&row.FirstName,
			&row.LastName,
			&row.Image,
			&row.Title,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		tasks = append(tasks, row)
	}

	return tasks, nil
}

func (q *ProjectStorePostgres) CountTasksByProjectId(
	ctx context.Context,
	projectId int64,
	arg *params.TaskSearch) (int64, error) {

	var query strings.Builder

	query.WriteString(`
	SELECT COUNT(t.id)
	FROM tasks t
	JOIN task_statuses ts ON ts.id = t.status
	JOIN task_priorities tp ON tp.id = t.priority
	WHERE t.project_id = $1
	`)

	queryArgs := []any{projectId}

	paramCount := 1

	if len(arg.Keyword) != 0 {
		paramCount++
		query.WriteString(" AND t.name LIKE $")
		query.WriteString(strconv.Itoa(paramCount))
		queryArgs = append(queryArgs, "%"+arg.Keyword+"%")
	}

	if len(arg.Status) != 0 {
		paramCount++
		query.WriteString(" AND ts.name = $")
		query.WriteString(strconv.Itoa(paramCount))
		queryArgs = append(queryArgs, arg.Status)
	}

	if len(arg.Priority) != 0 {
		paramCount++
		query.WriteString(" AND tp.name = $")
		query.WriteString(strconv.Itoa(paramCount))
		queryArgs = append(queryArgs, arg.Priority)
	}

	var count int64
	err := q.db.QueryRowContext(ctx, query.String(), queryArgs...).Scan(&count)
	if err != nil {
		return 0, store.ErrQueryFailed
	}

	return count, nil
}

func (q *ProjectStorePostgres) GetById(ctx context.Context, projectId int64) (*agg.Project, error) {

	query := `
	SELECT 
		p.id,
		p.name,
		p.created_at,
		ps.name AS status
	FROM projects p
	JOIN project_statuses ps ON ps.id = p.status
	WHERE p.id = $1
	`

	var project agg.Project
	err := q.db.QueryRowContext(ctx, query, projectId).Scan(
		&project.Id,
		&project.Name,
		&project.CreatedAt,
		&project.Status,
	)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	return &project, nil
}

func (q *ProjectStorePostgres) GetMembersByProjectId(
	ctx context.Context,
	projectId int64) ([]agg.ProjectUser, error) {

	query := `
	SELECT 
		u.id,
		u.first_name,
		u.last_name,
		r.name AS role,
		u.image,
		u.title
	FROM project_users pu
	JOIN users u ON u.id = pu.user_id
	JOIN roles r ON r.id = u.role
	WHERE pu.project_id = $1
	`

	rows, err := q.db.QueryContext(ctx, query, projectId)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	resp := make([]agg.ProjectUser, 0)

	for rows.Next() {
		var row agg.ProjectUser

		err := rows.Scan(
			&row.Id,
			&row.FirstName,
			&row.LastName,
			&row.Role,
			&row.Image,
			&row.Title,
		)
		if err != nil {
			return nil, store.ErrQueryFailed
		}

		resp = append(resp, row)
	}

	return resp, nil
}

func (q *ProjectStorePostgres) GetTaskMetricsByProjectId(
	ctx context.Context,
	projectId int64,
	status params.TaskStatus) ([]agg.TaskMetric, error) {

	query := `
	WITH RECURSIVE last_two_weeks(day) AS (
    	SELECT CURRENT_DATE
    	UNION ALL
    	SELECT day - INTERVAL '1 day'
    	FROM last_two_weeks
    	WHERE day > CURRENT_DATE - INTERVAL '13 days'
	)
	SELECT 
    	ltw.day,
    	COUNT(*) FILTER (WHERE t.project_id = $1 AND ts.name = $2) AS completed_count
	FROM last_two_weeks ltw
	LEFT JOIN tasks t ON DATE(t.updated_at) = ltw.day
	LEFT JOIN task_statuses ts ON ts.id = t.status
	GROUP BY ltw.day
	ORDER BY ltw.day
	`

	rows, err := q.db.QueryContext(ctx, query, projectId, status)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	resp := make([]agg.TaskMetric, 0)

	for rows.Next() {
		var row agg.TaskMetric
		if err := rows.Scan(&row.Key, &row.Value); err != nil {
			return nil, store.ErrQueryFailed
		}

		resp = append(resp, row)
	}

	return resp, nil
}

func (q *ProjectStorePostgres) GetCreatedMetricsByUserId(
	ctx context.Context,
	userId int64) ([]agg.ProjectMetric, error) {

	query := `
	WITH RECURSIVE months AS (
  		SELECT to_char(date_trunc('month', current_date), 'YYYY-MM') AS year_month,
         	date_trunc('month', current_date)::date AS start_date
  		UNION ALL
  		SELECT to_char(start_date - interval '1 month', 'YYYY-MM'),
        	(start_date - interval '1 month')::date
  		FROM months
  		WHERE start_date > (current_date - interval '11 months')
	)
	SELECT 
		m.year_month,
		COUNT(*) FILTER( WHERE pu.user_id = ? ) AS completed_count
	FROM months m
	LEFT JOIN projects p ON to_char(date_trunc('month', p.created_at), 'YYYY-MM') = m.year_month
	LEFT JOIN project_users pu ON pu.project_id = p.id
	GROUP BY m.year_month
	ORDER BY m.year_month
	`

	rows, err := q.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	result := make([]agg.ProjectMetric, 0)

	for rows.Next() {
		var row agg.ProjectMetric
		if err := rows.Scan(&row.Key, &row.Value); err != nil {
			return nil, store.ErrQueryFailed
		}

		result = append(result, row)
	}

	return result, nil
}

func (q *ProjectStorePostgres) SetStatusById(
	ctx context.Context,
	projectId int64,
	status params.ProjectStatus) error {

	query := `
	UPDATE projects
	SET status = (SELECT id FROM project_statuses WHERE name = $1)
	WHERE id = $2
	`

	_, err := q.db.ExecContext(ctx, query, status, projectId)
	if err != nil {
		return store.ErrUpdateFailed
	}

	return nil
}

func (q *ProjectStorePostgres) DeleteById(ctx context.Context, projectId int64) error {

	query := `
	DELETE FROM projects
	WHERE id = ?
	`

	_, err := q.db.ExecContext(ctx, query, projectId)
	if err != nil {
		return store.ErrDeleteFailed
	}

	return nil
}
