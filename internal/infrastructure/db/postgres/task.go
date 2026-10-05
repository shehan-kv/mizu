package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	"mizu/internal/domain/task"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}

func (r *TaskRepository) executor(ctx context.Context) executor {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (r *TaskRepository) Add(ctx context.Context, t *task.Task) error {
	ex := r.executor(ctx)

	_, err := ex.ExecContext(
		ctx,
		`INSERT INTO tasks(
			id,
			project_id,
			priority,
			status,
			name,
			description,
			estimated_minutes,
			version,
			created_at,
			updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID().String(),
		t.ProjectID().String(),
		t.Priority().String(),
		t.Status().String(),
		t.Name().String(),
		t.Description(),
		t.EstimatedMinutes().Int(),
		t.Version(),
		t.CreatedAt(),
		t.UpdatedAt(),
	)
	if err != nil {
		if pqErr, ok := errors.AsType[*pq.Error](err); ok {
			// 23505 = unique_violation
			if pqErr.Code == "23505" {
				return fmt.Errorf(
					"task.TaskRepository.Add: duplicate task id: %w",
					err,
				)
			}
		}

		return fmt.Errorf("task.TaskRepository.Add: %w", err)
	}

	assignees := t.Assignees()
	if len(assignees) == 0 {
		return nil
	}

	var sb strings.Builder
	args := make([]any, 0, len(assignees)*2)

	sb.WriteString(`INSERT INTO task_assignees(task_id, user_id) VALUES `)

	for i := range assignees {
		if i > 0 {
			sb.WriteString(", ")
		}

		base := len(args)

		sb.WriteString("(")
		sb.WriteString("$")
		sb.WriteString(strconv.Itoa(base + 1))
		sb.WriteString(", $")
		sb.WriteString(strconv.Itoa(base + 2))
		sb.WriteString(")")

		args = append(args, t.ID().String(), assignees[i].String())
	}

	_, err = ex.ExecContext(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf("task.TaskRepository.Add: %w", err)
	}

	return nil
}

func (r *TaskRepository) Get(ctx context.Context, tID task.TaskID) (*task.Task, error) {
	ex := r.executor(ctx)

	var (
		rawID            string
		rawProjectID     string
		priority         string
		status           string
		name             string
		description      string
		estimatedMinutes int
		version          int
		createdAt        time.Time
		updatedAt        time.Time
	)

	err := ex.QueryRowContext(
		ctx,
		`SELECT
			id,
			project_id,
			priority,
			status,
			name,
			description,
			estimated_minutes,
			version,
			created_at,
			updated_at
		FROM tasks
		WHERE id = $1`,
		tID.String(),
	).Scan(
		&rawID,
		&rawProjectID,
		&priority,
		&status,
		&name,
		&description,
		&estimatedMinutes,
		&version,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf(
				"%w: %w",
				task.ErrTaskNotFound,
				err,
			)
		}

		return nil, fmt.Errorf("task.TaskRepository.Get: %w", err)
	}

	rows, err := ex.QueryContext(
		ctx,
		`SELECT user_id FROM task_assignees WHERE task_id = $1`,
		rawID,
	)
	if err != nil {
		return nil, fmt.Errorf("task.TaskRepository.Get: %w", err)
	}
	defer rows.Close()

	assignees := make([]iam.UserID, 0)

	for rows.Next() {
		var rawUserID string

		if err := rows.Scan(&rawUserID); err != nil {
			return nil, fmt.Errorf("task.TaskRepository.Get: %w", err)
		}

		assignees = append(assignees, iam.UserID(rawUserID))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("task.TaskRepository.Get: %w", err)
	}

	taskID, err := task.NewTaskID(rawID)
	if err != nil {
		return nil, fmt.Errorf("task.TaskRepository.Get: %w", err)
	}

	projectID, err := project.NewProjectID(rawProjectID)
	if err != nil {
		return nil, fmt.Errorf("task.TaskRepository.Get: %w", err)
	}

	priorityVO, err := task.NewPriority(priority)
	if err != nil {
		return nil, fmt.Errorf("task.TaskRepository.Get: %w", err)
	}

	statusVO, err := task.NewStatus(status)
	if err != nil {
		return nil, fmt.Errorf("task.TaskRepository.Get: %w", err)
	}

	nameVO, err := task.NewName(name)
	if err != nil {
		return nil, fmt.Errorf("task.TaskRepository.Get: %w", err)
	}

	minutesVO, err := task.NewMinutes(estimatedMinutes)
	if err != nil {
		return nil, fmt.Errorf("task.TaskRepository.Get: %w", err)
	}

	return task.RestoreTask(
		taskID,
		projectID,
		priorityVO,
		statusVO,
		nameVO,
		description,
		minutesVO,
		assignees,
		version,
		createdAt,
		updatedAt,
	), nil
}

func (r *TaskRepository) ListCompletedPerDay(ctx context.Context, pID project.ProjectID) ([]task.Metric, error) {

	rows, err := r.executor(ctx).QueryContext(
		ctx,
		`WITH RECURSIVE last_two_weeks(day) AS (
			SELECT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date

			UNION ALL

			SELECT day - 1
			FROM last_two_weeks
			WHERE day > (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date - 13
		)
		SELECT
			TO_CHAR(ltw.day, 'YYYY-MM-DD') AS day,
			COUNT(t.id) AS completed_count
		FROM last_two_weeks ltw
		LEFT JOIN tasks t
			ON (t.updated_at AT TIME ZONE 'UTC')::date = ltw.day
			AND t.project_id = $1
			AND t.status = $2
		GROUP BY ltw.day
		ORDER BY ltw.day`,
		pID.String(),
		task.StatusCompleted.String(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"task.TaskRepository.ListCompletedPerDay: %w",
			err,
		)
	}
	defer rows.Close()

	metrics := make([]task.Metric, 0)

	for rows.Next() {
		var key string
		var value int64

		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf(
				"task.TaskRepository.ListCompletedPerDay: %w",
				err,
			)
		}

		metrics = append(metrics, task.NewMetric(key, value))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"task.TaskRepository.ListCompletedPerDay: %w",
			err,
		)
	}

	return metrics, nil
}

func (r *TaskRepository) GetStatsByProject(ctx context.Context, pID project.ProjectID) (task.Stats, error) {

	ex := r.executor(ctx)

	var (
		total     int
		completed int
	)

	err := ex.QueryRowContext(
		ctx,
		`SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (
				WHERE status = $1
			) AS completed
		FROM tasks
		WHERE project_id = $2`,
		task.StatusCompleted.String(),
		pID.String(),
	).Scan(
		&total,
		&completed,
	)
	if err != nil {
		return task.Stats{}, fmt.Errorf(
			"task.TaskRepository.GetStatsByProject: %w",
			err,
		)
	}

	return task.NewStats(total, completed), nil
}

func (r *TaskRepository) List(ctx context.Context, f task.TaskFilter, p common.Page) ([]*task.Task, error) {
	ex := r.executor(ctx)

	var query strings.Builder

	args := make([]any, 0, 10)

	query.WriteString(`
        SELECT
            id,
            project_id,
            priority,
            status,
            name,
            description,
            estimated_minutes,
            version,
            created_at,
            updated_at
        FROM tasks
        WHERE project_id = $1
    `)

	args = append(args, f.ProjectID.String())

	argPos := 2

	if f.Keyword != nil {
		query.WriteString(" AND (name ILIKE $")
		query.WriteString(strconv.Itoa(argPos))
		query.WriteString(" OR description ILIKE $")
		query.WriteString(strconv.Itoa(argPos + 1))
		query.WriteString(")")

		keyword := "%" + *f.Keyword + "%"

		args = append(args, keyword, keyword)
		argPos += 2
	}

	if f.Status != nil {
		query.WriteString(" AND status = $")
		query.WriteString(strconv.Itoa(argPos))

		args = append(args, f.Status.String())
		argPos++
	}

	if f.Priority != nil {
		query.WriteString(" AND priority = $")
		query.WriteString(strconv.Itoa(argPos))

		args = append(args, f.Priority.String())
		argPos++
	}

	query.WriteString(" ORDER BY created_at DESC LIMIT $")
	query.WriteString(strconv.Itoa(argPos))
	query.WriteString(" OFFSET $")
	query.WriteString(strconv.Itoa(argPos + 1))

	args = append(args, p.Limit(), p.Offset())

	rows, err := ex.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("task.TaskRepository.List: %w", err)
	}
	defer rows.Close()

	type rawTask struct {
		id               string
		projectID        string
		priority         string
		status           string
		name             string
		description      string
		estimatedMinutes int
		version          int
		createdAt        time.Time
		updatedAt        time.Time
	}

	rawTasks := make([]rawTask, 0)
	taskIDs := make([]string, 0)

	for rows.Next() {
		var rt rawTask

		if err := rows.Scan(
			&rt.id,
			&rt.projectID,
			&rt.priority,
			&rt.status,
			&rt.name,
			&rt.description,
			&rt.estimatedMinutes,
			&rt.version,
			&rt.createdAt,
			&rt.updatedAt,
		); err != nil {
			return nil, fmt.Errorf("task.TaskRepository.List: %w", err)
		}

		rawTasks = append(rawTasks, rt)
		taskIDs = append(taskIDs, rt.id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("task.TaskRepository.List: %w", err)
	}

	assigneesByTaskID := make(map[string][]iam.UserID)

	if len(taskIDs) > 0 {
		assigneeRows, err := ex.QueryContext(
			ctx,
			`
            SELECT task_id, user_id
            FROM task_assignees
            WHERE task_id = ANY($1)
            `,
			pq.Array(taskIDs),
		)
		if err != nil {
			return nil, fmt.Errorf("task.TaskRepository.List: %w", err)
		}
		defer assigneeRows.Close()

		for assigneeRows.Next() {
			var taskID, userID string

			if err := assigneeRows.Scan(&taskID, &userID); err != nil {
				return nil, fmt.Errorf("task.TaskRepository.List: %w", err)
			}

			assigneesByTaskID[taskID] = append(
				assigneesByTaskID[taskID],
				iam.UserID(userID),
			)
		}

		if err := assigneeRows.Err(); err != nil {
			return nil, fmt.Errorf("task.TaskRepository.List: %w", err)
		}
	}

	tasks := make([]*task.Task, 0, len(rawTasks))

	for _, rt := range rawTasks {
		taskID, err := task.NewTaskID(rt.id)
		if err != nil {
			return nil, fmt.Errorf("task.TaskRepository.List: %w", err)
		}

		projectID, err := project.NewProjectID(rt.projectID)
		if err != nil {
			return nil, fmt.Errorf("task.TaskRepository.List: %w", err)
		}

		priorityVO, err := task.NewPriority(rt.priority)
		if err != nil {
			return nil, fmt.Errorf("task.TaskRepository.List: %w", err)
		}

		statusVO, err := task.NewStatus(rt.status)
		if err != nil {
			return nil, fmt.Errorf("task.TaskRepository.List: %w", err)
		}

		nameVO, err := task.NewName(rt.name)
		if err != nil {
			return nil, fmt.Errorf("task.TaskRepository.List: %w", err)
		}

		minutesVO, err := task.NewMinutes(rt.estimatedMinutes)
		if err != nil {
			return nil, fmt.Errorf("task.TaskRepository.List: %w", err)
		}

		tasks = append(tasks, task.RestoreTask(
			taskID,
			projectID,
			priorityVO,
			statusVO,
			nameVO,
			rt.description,
			minutesVO,
			assigneesByTaskID[rt.id],
			rt.version,
			rt.createdAt,
			rt.updatedAt,
		))
	}

	return tasks, nil
}

func (r *TaskRepository) ListStatsByProjects(ctx context.Context, projectIDs []project.ProjectID) (map[project.ProjectID]task.Stats, error) {

	ex := r.executor(ctx)

	statsByProject := make(map[project.ProjectID]task.Stats)

	if len(projectIDs) == 0 {
		return statsByProject, nil
	}

	var query strings.Builder

	query.WriteString(`
        SELECT
            project_id,
            COUNT(*) AS total,
            COUNT(*) FILTER (WHERE status = $1) AS completed
        FROM tasks
        WHERE project_id = ANY($2)
        GROUP BY project_id
    `)

	args := make([]any, 0, 2)
	args = append(args,
		task.StatusCompleted.String(),
		pq.Array(projectIDs),
	)

	rows, err := ex.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf(
			"task.TaskRepository.ListStatsByProjects: %w",
			err,
		)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			rawProjectID string
			total        int
			completed    int
		)

		if err := rows.Scan(&rawProjectID, &total, &completed); err != nil {
			return nil, fmt.Errorf(
				"task.TaskRepository.ListStatsByProjects: %w",
				err,
			)
		}

		pID, err := project.NewProjectID(rawProjectID)
		if err != nil {
			return nil, fmt.Errorf(
				"task.TaskRepository.ListStatsByProjects: %w",
				err,
			)
		}

		statsByProject[pID] = task.NewStats(total, completed)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"task.TaskRepository.ListStatsByProjects: %w",
			err,
		)
	}

	return statsByProject, nil
}

func (r *TaskRepository) Count(ctx context.Context, f task.TaskFilter) (int, error) {
	ex := r.executor(ctx)

	var query strings.Builder

	query.WriteString(`
        SELECT COUNT(*)
        FROM tasks
        WHERE project_id = $1
    `)

	args := make([]any, 0, 5)
	args = append(args, f.ProjectID.String())

	argPos := 2

	if f.Keyword != nil {
		query.WriteString(" AND (name ILIKE $")
		query.WriteString(strconv.Itoa(argPos))
		query.WriteString(" OR description ILIKE $")
		query.WriteString(strconv.Itoa(argPos + 1))
		query.WriteString(")")

		keyword := "%" + *f.Keyword + "%"

		args = append(args, keyword, keyword)
		argPos += 2
	}

	if f.Status != nil {
		query.WriteString(" AND status = $")
		query.WriteString(strconv.Itoa(argPos))

		args = append(args, f.Status.String())
		argPos++
	}

	if f.Priority != nil {
		query.WriteString(" AND priority = $")
		query.WriteString(strconv.Itoa(argPos))

		args = append(args, f.Priority.String())
		argPos++
	}

	var count int

	err := ex.QueryRowContext(ctx, query.String(), args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("task.TaskRepository.Count: %w", err)
	}

	return count, nil
}

func (r *TaskRepository) Save(ctx context.Context, t *task.Task) error {
	ex := r.executor(ctx)

	res, err := ex.ExecContext(
		ctx,
		`UPDATE tasks
		 SET priority = $1,
		     status = $2,
		     name = $3,
		     description = $4,
		     estimated_minutes = $5,
		     version = version + 1,
		     updated_at = $6
		 WHERE id = $7 AND version = $8`,
		t.Priority().String(),
		t.Status().String(),
		t.Name().String(),
		t.Description(),
		t.EstimatedMinutes().Int(),
		t.UpdatedAt(),
		t.ID().String(),
		t.Version(),
	)
	if err != nil {
		return fmt.Errorf("task.TaskRepository.Save: update task: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("task.TaskRepository.Save: rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task.TaskRepository.Save: %w", task.ErrTaskConcurrentModification)
	}

	_, err = ex.ExecContext(
		ctx,
		`DELETE FROM task_assignees WHERE task_id = $1`,
		t.ID().String(),
	)
	if err != nil {
		return fmt.Errorf("task.TaskRepository.Save: delete assignees: %w", err)
	}

	assignees := t.Assignees()
	if len(assignees) == 0 {
		return nil
	}

	var query strings.Builder

	query.WriteString(`
			INSERT INTO task_assignees (task_id, user_id) VALUES
		`)

	args := make([]any, 0, len(assignees)*2)

	for i, userID := range assignees {
		if i > 0 {
			query.WriteString(",")
		}

		base := i*2 + 1

		query.WriteString("(")
		query.WriteString("$")
		query.WriteString(strconv.Itoa(base))
		query.WriteString(", $")
		query.WriteString(strconv.Itoa(base + 1))
		query.WriteString(")")

		args = append(args,
			t.ID().String(),
			userID.String(),
		)
	}

	_, err = ex.ExecContext(ctx, query.String(), args...)
	if err != nil {
		return fmt.Errorf("task.TaskRepository.Save: insert assignees: %w", err)
	}

	return nil
}

func (r *TaskRepository) Remove(ctx context.Context, t *task.Task) error {
	ex := r.executor(ctx)

	res, err := ex.ExecContext(
		ctx,
		`DELETE FROM tasks
		 WHERE id = $1 AND version = $2`,
		t.ID().String(),
		t.Version(),
	)
	if err != nil {
		return fmt.Errorf("task.TaskRepository.Remove: delete task: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("task.TaskRepository.Remove: rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf(
			"task.TaskRepository.Remove: %w",
			task.ErrTaskConcurrentModification,
		)
	}

	return nil
}
