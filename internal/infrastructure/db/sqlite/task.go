package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	"mizu/internal/domain/task"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
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
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
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
		if sqlite3Err, ok := errors.AsType[sqlite3.Error](err); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return fmt.Errorf(
					"task.TaskRepository.Add: duplicate task id: %w",
					err,
				)
			}
		}

		return fmt.Errorf("task.TaskRepository.Add: %w", err)
	}

	assignees := t.Assignees()

	var sb strings.Builder
	args := make([]any, 0, len(assignees)*2)

	sb.WriteString("INSERT INTO task_assignees(task_id, user_id) VALUES ")

	for i, assigneeID := range assignees {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString("(?, ?)")
		args = append(args, t.ID().String(), assigneeID.String())
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
        WHERE id = ?`,
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
		`SELECT user_id FROM task_assignees WHERE task_id = ?`,
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
		`WITH last_two_weeks(day) AS (
			SELECT date('now')
			UNION ALL
			SELECT date(day, '-1 day')
			FROM last_two_weeks
			WHERE day > date('now', '-13 days')
		)
		SELECT
			ltw.day,
			COUNT(*) FILTER (
				WHERE t.project_id = ? AND t.status = ?
			) AS completed_count
		FROM last_two_weeks ltw
		LEFT JOIN tasks t ON date(t.updated_at) = ltw.day
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
                WHERE status = ?
            ) AS completed
        FROM tasks
        WHERE project_id = ?`,
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

	stats := task.NewStats(
		total,
		completed,
	)

	return stats, nil

}

func (r *TaskRepository) List(ctx context.Context, f task.TaskFilter, p common.Page) ([]*task.Task, error) {
	ex := r.executor(ctx)

	var query strings.Builder

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
        WHERE project_id = ?
    `)

	args := make([]any, 0, 7)
	args = append(args, f.ProjectID.String())

	if f.Keyword != nil {

		query.WriteString(`
            AND (
                name LIKE ?
                OR description LIKE ?
            )
        `)

		keyword := "%" + *f.Keyword + "%"

		args = append(
			args,
			keyword,
			keyword,
		)
	}

	if f.Status != nil {

		query.WriteString(` AND status = ?`)

		args = append(
			args,
			f.Status.String(),
		)
	}

	if f.Priority != nil {

		query.WriteString(` AND priority = ?`)

		args = append(
			args,
			f.Priority.String(),
		)
	}

	query.WriteString(`
        ORDER BY created_at DESC
        LIMIT ?
        OFFSET ?
    `)

	args = append(
		args,
		p.Limit(),
		p.Offset(),
	)

	rows, err := ex.QueryContext(
		ctx,
		query.String(),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"task.TaskRepository.List: %w",
			err,
		)
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
			return nil, fmt.Errorf(
				"task.TaskRepository.List: %w",
				err,
			)
		}

		rawTasks = append(rawTasks, rt)
		taskIDs = append(taskIDs, rt.id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"task.TaskRepository.List: %w",
			err,
		)
	}

	assigneesByTaskID := make(map[string][]iam.UserID)

	if len(taskIDs) > 0 {

		assigneeArgs := make([]any, len(taskIDs))

		var assigneeQuery strings.Builder

		assigneeQuery.WriteString(`
            SELECT
                task_id,
                user_id
            FROM task_assignees
            WHERE task_id IN (
        `)

		for i := range taskIDs {

			if i > 0 {
				assigneeQuery.WriteString(",")
			}

			assigneeQuery.WriteString("?")

			assigneeArgs[i] = taskIDs[i]
		}

		assigneeQuery.WriteString(")")

		assigneeRows, err := ex.QueryContext(
			ctx,
			assigneeQuery.String(),
			assigneeArgs...,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"task.TaskRepository.List: %w",
				err,
			)
		}
		defer assigneeRows.Close()

		for assigneeRows.Next() {
			var (
				rawTaskID string
				rawUserID string
			)

			if err := assigneeRows.Scan(
				&rawTaskID,
				&rawUserID,
			); err != nil {
				return nil, fmt.Errorf(
					"task.TaskRepository.List: %w",
					err,
				)
			}

			assigneesByTaskID[rawTaskID] = append(
				assigneesByTaskID[rawTaskID],
				iam.UserID(rawUserID),
			)
		}

		if err := assigneeRows.Err(); err != nil {
			return nil, fmt.Errorf(
				"task.TaskRepository.List: %w",
				err,
			)
		}
	}

	tasks := make([]*task.Task, 0, len(rawTasks))

	for i := range rawTasks {

		rt := rawTasks[i]

		taskID, err := task.NewTaskID(rt.id)
		if err != nil {
			return nil, fmt.Errorf(
				"task.TaskRepository.List: %w",
				err,
			)
		}

		projectID, err := project.NewProjectID(rt.projectID)
		if err != nil {
			return nil, fmt.Errorf(
				"task.TaskRepository.List: %w",
				err,
			)
		}

		priorityVO, err := task.NewPriority(rt.priority)
		if err != nil {
			return nil, fmt.Errorf(
				"task.TaskRepository.List: %w",
				err,
			)
		}

		statusVO, err := task.NewStatus(rt.status)
		if err != nil {
			return nil, fmt.Errorf(
				"task.TaskRepository.List: %w",
				err,
			)
		}

		nameVO, err := task.NewName(rt.name)
		if err != nil {
			return nil, fmt.Errorf(
				"task.TaskRepository.List: %w",
				err,
			)
		}

		minutesVO, err := task.NewMinutes(rt.estimatedMinutes)
		if err != nil {
			return nil, fmt.Errorf(
				"task.TaskRepository.List: %w",
				err,
			)
		}

		tasks = append(
			tasks,
			task.RestoreTask(
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
			),
		)
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
            COUNT(*) FILTER (
                WHERE status = ?
            ) AS completed
        FROM tasks
        WHERE project_id IN (
    `)

	args := make([]any, 0, len(projectIDs)+1)
	args = append(args, task.StatusCompleted.String())

	for i := range projectIDs {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString("?")
		args = append(args, projectIDs[i].String())
	}

	query.WriteString(`
        )
        GROUP BY project_id
    `)

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

		if err := rows.Scan(
			&rawProjectID,
			&total,
			&completed,
		); err != nil {
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
        WHERE project_id = ?
    `)

	args := make([]any, 0, 5)
	args = append(args, f.ProjectID.String())

	if f.Keyword != nil {
		query.WriteString(`
            AND (
                name LIKE ?
                OR description LIKE ?
            )
        `)

		keyword := "%" + *f.Keyword + "%"

		args = append(
			args,
			keyword,
			keyword,
		)
	}

	if f.Status != nil {
		query.WriteString(` AND status = ?`)

		args = append(args, f.Status.String())
	}

	if f.Priority != nil {
		query.WriteString(` AND priority = ?`)

		args = append(args, f.Priority.String())
	}

	var count int

	err := ex.QueryRowContext(
		ctx,
		query.String(),
		args...,
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf(
			"task.TaskRepository.Count: %w",
			err,
		)
	}

	return count, nil
}

func (r *TaskRepository) Save(ctx context.Context, t *task.Task) error {

	ex := r.executor(ctx)

	res, err := ex.ExecContext(
		ctx,
		`UPDATE tasks
		 SET priority = ?,
		     status = ?,
		     name = ?,
		     description = ?,
		     estimated_minutes = ?,
		     version = version + 1,
		     updated_at = ?
		 WHERE id = ? AND version = ?`,
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

	// Delete previous assignees
	_, err = ex.ExecContext(
		ctx,
		`DELETE FROM task_assignees WHERE task_id = ?`,
		t.ID().String(),
	)
	if err != nil {
		return fmt.Errorf("task.TaskRepository.Save: delete assignees: %w", err)
	}

	// Add new assignees
	assignees := t.Assignees()

	if len(assignees) > 0 {
		var sb strings.Builder

		sb.WriteString(`
			INSERT INTO task_assignees (task_id, user_id)
			VALUES
		`)

		args := make([]any, 0, len(assignees)*2)

		for i, userID := range assignees {
			if i > 0 {
				sb.WriteString(",")
			}

			sb.WriteString("(?, ?)")
			args = append(args,
				t.ID().String(),
				userID.String(),
			)
		}

		_, err = ex.ExecContext(ctx, sb.String(), args...)
		if err != nil {
			return fmt.Errorf("task.TaskRepository.Save: insert assignees: %w", err)
		}
	}

	return nil
}

func (r *TaskRepository) Remove(ctx context.Context, t *task.Task) error {
	ex := r.executor(ctx)

	res, err := ex.ExecContext(
		ctx,
		`DELETE FROM tasks
		 WHERE id = ? AND version = ?`,
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
