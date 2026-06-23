package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	"strconv"
	"strings"
	"time"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (r *ProjectRepository) executor(ctx context.Context) executor {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (r *ProjectRepository) Add(ctx context.Context, p *project.Project) error {
	ex := r.executor(ctx)

	_, err := ex.ExecContext(
		ctx,
		`INSERT INTO projects(
			id,
			status,
			name,
			version,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		p.ID().String(),
		p.Status().String(),
		p.Name().String(),
		p.Version(),
		p.CreatedAt(),
		p.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf(
			"project.ProjectRepository.Add: %w",
			err,
		)
	}

	members := p.Members()

	var sb strings.Builder
	args := make([]any, 0, len(members)*2)

	sb.WriteString(`
		INSERT INTO project_members(
			project_id,
			user_id
		) VALUES 
	`)

	argPos := 1

	for i, memberID := range members {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString("($")
		sb.WriteString(strconv.Itoa(argPos))
		sb.WriteString(", $")
		sb.WriteString(strconv.Itoa(argPos + 1))
		sb.WriteString(")")

		args = append(
			args,
			p.ID().String(),
			memberID.String(),
		)

		argPos += 2
	}

	_, err = ex.ExecContext(
		ctx,
		sb.String(),
		args...,
	)
	if err != nil {
		return fmt.Errorf(
			"project.ProjectRepository.Add: %w",
			err,
		)
	}

	return nil
}

func (r *ProjectRepository) Get(ctx context.Context, id project.ProjectID) (*project.Project, error) {
	ex := r.executor(ctx)

	var (
		rawID     string
		name      string
		status    string
		version   int
		createdAt time.Time
		updatedAt time.Time
	)

	err := ex.QueryRowContext(
		ctx,
		`SELECT
			id,
			name,
			status,
			version,
			created_at,
			updated_at
		FROM projects
		WHERE id = $1`,
		id.String(),
	).Scan(
		&rawID,
		&name,
		&status,
		&version,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", project.ErrProjectNotFound, err)
		}

		return nil, fmt.Errorf("project.ProjectRepository.Get: %w", err)
	}

	rows, err := ex.QueryContext(
		ctx,
		`SELECT user_id
		FROM project_members
		WHERE project_id = $1`,
		rawID,
	)
	if err != nil {
		return nil, fmt.Errorf("project.ProjectRepository.Get: %w", err)
	}
	defer rows.Close()

	members := make([]iam.UserID, 0)

	for rows.Next() {
		var rawUserID string

		if err := rows.Scan(&rawUserID); err != nil {
			return nil, fmt.Errorf("project.ProjectRepository.Get: %w", err)
		}

		members = append(members, iam.UserID(rawUserID))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("project.ProjectRepository.Get: %w", err)
	}

	projectID, err := project.NewProjectID(rawID)
	if err != nil {
		return nil, fmt.Errorf("project.ProjectRepository.Get: %w", err)
	}

	nameVO, err := project.NewName(name)
	if err != nil {
		return nil, fmt.Errorf("project.ProjectRepository.Get: %w", err)
	}

	statusVO, err := project.NewStatus(status)
	if err != nil {
		return nil, fmt.Errorf("project.ProjectRepository.Get: %w", err)
	}

	return project.RestoreProject(
		projectID,
		nameVO,
		statusVO,
		members,
		version,
		createdAt,
		updatedAt,
	), nil
}

func (r *ProjectRepository) List(ctx context.Context, f project.Filter, p common.Page) ([]*project.Project, error) {
	ex := r.executor(ctx)

	var (
		sb     strings.Builder
		args   = make([]any, 0, 5)
		argPos = 1
	)

	sb.WriteString(`
		SELECT
			p.id,
			p.name,
			p.status,
			p.version,
			p.created_at,
			p.updated_at
		FROM projects p
		INNER JOIN project_members pm
			ON p.id = pm.project_id
		WHERE pm.user_id = $
	`)
	sb.WriteString(strconv.Itoa(argPos))

	args = append(args, f.MemberID.String())
	argPos++

	if f.Keyword != nil {
		sb.WriteString(` AND p.name ILIKE $`)
		sb.WriteString(strconv.Itoa(argPos))

		args = append(args, "%"+*f.Keyword+"%")
		argPos++
	}

	if f.Status != nil {
		sb.WriteString(` AND p.status = $`)
		sb.WriteString(strconv.Itoa(argPos))

		args = append(args, f.Status.String())
		argPos++
	}

	sb.WriteString(` ORDER BY p.created_at DESC LIMIT $`)
	sb.WriteString(strconv.Itoa(argPos))

	args = append(args, p.Limit())
	argPos++

	sb.WriteString(` OFFSET $`)
	sb.WriteString(strconv.Itoa(argPos))

	args = append(args, p.Offset())

	rows, err := ex.QueryContext(
		ctx,
		sb.String(),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"project.ProjectRepository.List: %w",
			err,
		)
	}
	defer rows.Close()

	type projectRow struct {
		rawID     string
		name      string
		status    string
		version   int
		createdAt time.Time
		updatedAt time.Time
	}

	projectRows := make([]projectRow, 0)
	rawIDs := make([]any, 0)

	for rows.Next() {
		var row projectRow

		if err := rows.Scan(
			&row.rawID,
			&row.name,
			&row.status,
			&row.version,
			&row.createdAt,
			&row.updatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"project.ProjectRepository.List: %w",
				err,
			)
		}

		projectRows = append(projectRows, row)
		rawIDs = append(rawIDs, row.rawID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"project.ProjectRepository.List: %w",
			err,
		)
	}

	if len(projectRows) == 0 {
		return nil, nil
	}

	var msb strings.Builder

	msb.WriteString(`
		SELECT
			project_id,
			user_id
		FROM project_members
		WHERE project_id IN (
	`)

	argPos = 1

	for i := range rawIDs {
		if i > 0 {
			msb.WriteString(", ")
		}

		msb.WriteString("$")
		msb.WriteString(strconv.Itoa(argPos))

		argPos++
	}

	msb.WriteString(")")

	memberRows, err := ex.QueryContext(
		ctx,
		msb.String(),
		rawIDs...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"project.ProjectRepository.List: %w",
			err,
		)
	}
	defer memberRows.Close()

	memberMap := make(map[string][]iam.UserID)

	for memberRows.Next() {
		var (
			projectID string
			userID    string
		)

		if err := memberRows.Scan(
			&projectID,
			&userID,
		); err != nil {
			return nil, fmt.Errorf(
				"project.ProjectRepository.List: %w",
				err,
			)
		}

		memberMap[projectID] = append(
			memberMap[projectID],
			iam.UserID(userID),
		)
	}

	if err := memberRows.Err(); err != nil {
		return nil, fmt.Errorf(
			"project.ProjectRepository.List: %w",
			err,
		)
	}

	projects := make(
		[]*project.Project,
		0,
		len(projectRows),
	)

	for _, row := range projectRows {
		projectID, err := project.NewProjectID(row.rawID)
		if err != nil {
			return nil, fmt.Errorf(
				"project.ProjectRepository.List: %w",
				err,
			)
		}

		nameVO, err := project.NewName(row.name)
		if err != nil {
			return nil, fmt.Errorf(
				"project.ProjectRepository.List: %w",
				err,
			)
		}

		statusVO, err := project.NewStatus(row.status)
		if err != nil {
			return nil, fmt.Errorf(
				"project.ProjectRepository.List: %w",
				err,
			)
		}

		projects = append(
			projects,
			project.RestoreProject(
				projectID,
				nameVO,
				statusVO,
				memberMap[row.rawID],
				row.version,
				row.createdAt,
				row.updatedAt,
			),
		)
	}

	return projects, nil
}

func (r *ProjectRepository) ListByIDs(ctx context.Context, pIDs []project.ProjectID) ([]*project.Project, error) {
	if len(pIDs) == 0 {
		return nil, nil
	}

	ex := r.executor(ctx)

	// Fetch members first
	var msb strings.Builder

	args := make([]any, 0, len(pIDs))
	argPos := 1

	msb.WriteString(`
		SELECT
			project_id,
			user_id
		FROM project_members
		WHERE project_id IN (
	`)

	for i, id := range pIDs {
		if i > 0 {
			msb.WriteString(", ")
		}

		msb.WriteString("$")
		msb.WriteString(strconv.Itoa(argPos))

		args = append(args, id.String())
		argPos++
	}

	msb.WriteString(")")

	memberRows, err := ex.QueryContext(
		ctx,
		msb.String(),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("project.ProjectRepository.ListByIDs: %w", err)
	}

	defer memberRows.Close()

	memberMap := make(map[string][]iam.UserID)

	for memberRows.Next() {
		var (
			projectID string
			userID    string
		)

		if err := memberRows.Scan(&projectID, &userID); err != nil {
			return nil, fmt.Errorf("project.ProjectRepository.ListByIDs: %w", err)
		}

		memberMap[projectID] = append(memberMap[projectID], iam.UserID(userID))
	}

	if err := memberRows.Err(); err != nil {
		return nil, fmt.Errorf("project.ProjectRepository.ListByIDs: %w", err)
	}

	// Fetch projects
	var sb strings.Builder

	args = make([]any, 0, len(pIDs))
	argPos = 1

	sb.WriteString(`
		SELECT
			id,
			name,
			status,
			version,
			created_at,
			updated_at
		FROM projects
		WHERE id IN (
	`)

	for i, id := range pIDs {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString("$")
		sb.WriteString(strconv.Itoa(argPos))

		args = append(args, id.String())
		argPos++
	}

	sb.WriteString(")")

	rows, err := ex.QueryContext(
		ctx,
		sb.String(),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"project.ProjectRepository.ListByIDs: %w",
			err,
		)
	}
	defer rows.Close()

	projects := make(
		[]*project.Project,
		0,
		len(pIDs),
	)

	for rows.Next() {
		var (
			rawID     string
			name      string
			status    string
			version   int
			createdAt time.Time
			updatedAt time.Time
		)

		if err := rows.Scan(
			&rawID,
			&name,
			&status,
			&version,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("project.ProjectRepository.ListByIDs: %w", err)
		}

		projectID, err := project.NewProjectID(rawID)
		if err != nil {
			return nil, fmt.Errorf("project.ProjectRepository.ListByIDs: %w", err)
		}

		nameVO, err := project.NewName(name)
		if err != nil {
			return nil, fmt.Errorf("project.ProjectRepository.ListByIDs: %w", err)
		}

		statusVO, err := project.NewStatus(status)
		if err != nil {
			return nil, fmt.Errorf("project.ProjectRepository.ListByIDs: %w", err)
		}

		projects = append(projects, project.RestoreProject(
			projectID,
			nameVO,
			statusVO,
			memberMap[rawID],
			version,
			createdAt,
			updatedAt,
		),
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("project.ProjectRepository.ListByIDs: %w", err)
	}

	if len(projects) != len(pIDs) {
		return nil, project.ErrProjectNotFound
	}

	return projects, nil
}

func (r *ProjectRepository) ListByMember(ctx context.Context, memberID iam.UserID) ([]*project.Project, error) {
	ex := r.executor(ctx)

	rows, err := ex.QueryContext(
		ctx,
		`SELECT
			p.id,
			p.name,
			p.status,
			p.version,
			p.created_at,
			p.updated_at
		FROM projects p
		INNER JOIN project_members pm
			ON p.id = pm.project_id
		WHERE pm.user_id = $1
		ORDER BY p.created_at DESC`,
		memberID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"project.ProjectRepository.ListByMember: %w",
			err,
		)
	}
	defer rows.Close()

	type projectRow struct {
		rawID     string
		name      string
		status    string
		version   int
		createdAt time.Time
		updatedAt time.Time
	}

	projectRows := make([]projectRow, 0)
	rawIDs := make([]any, 0)

	for rows.Next() {
		var row projectRow

		if err := rows.Scan(
			&row.rawID,
			&row.name,
			&row.status,
			&row.version,
			&row.createdAt,
			&row.updatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"project.ProjectRepository.ListByMember: %w",
				err,
			)
		}

		projectRows = append(projectRows, row)
		rawIDs = append(rawIDs, row.rawID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"project.ProjectRepository.ListByMember: %w",
			err,
		)
	}

	if len(projectRows) == 0 {
		return nil, nil
	}

	var msb strings.Builder

	msb.WriteString(`
		SELECT
			project_id,
			user_id
		FROM project_members
		WHERE project_id IN (
	`)

	argPos := 1

	for i := range rawIDs {
		if i > 0 {
			msb.WriteString(", ")
		}

		msb.WriteString("$")
		msb.WriteString(strconv.Itoa(argPos))

		argPos++
	}

	msb.WriteString(")")

	memberRows, err := ex.QueryContext(
		ctx,
		msb.String(),
		rawIDs...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"project.ProjectRepository.ListByMember: %w",
			err,
		)
	}
	defer memberRows.Close()

	memberMap := make(map[string][]iam.UserID)

	for memberRows.Next() {
		var (
			projectID string
			userID    string
		)

		if err := memberRows.Scan(
			&projectID,
			&userID,
		); err != nil {
			return nil, fmt.Errorf(
				"project.ProjectRepository.ListByMember: %w",
				err,
			)
		}

		memberMap[projectID] = append(
			memberMap[projectID],
			iam.UserID(userID),
		)
	}

	if err := memberRows.Err(); err != nil {
		return nil, fmt.Errorf(
			"project.ProjectRepository.ListByMember: %w",
			err,
		)
	}

	projects := make(
		[]*project.Project,
		0,
		len(projectRows),
	)

	for _, row := range projectRows {
		projectID, err := project.NewProjectID(row.rawID)
		if err != nil {
			return nil, fmt.Errorf(
				"project.ProjectRepository.ListByMember: %w",
				err,
			)
		}

		nameVO, err := project.NewName(row.name)
		if err != nil {
			return nil, fmt.Errorf(
				"project.ProjectRepository.ListByMember: %w",
				err,
			)
		}

		statusVO, err := project.NewStatus(row.status)
		if err != nil {
			return nil, fmt.Errorf(
				"project.ProjectRepository.ListByMember: %w",
				err,
			)
		}

		projects = append(
			projects,
			project.RestoreProject(
				projectID,
				nameVO,
				statusVO,
				memberMap[row.rawID],
				row.version,
				row.createdAt,
				row.updatedAt,
			),
		)
	}

	return projects, nil
}

func (r *ProjectRepository) ListCreatedPerDay(ctx context.Context, mID iam.UserID) ([]project.Metric, error) {
	rows, err := r.executor(ctx).QueryContext(
		ctx,
		`
		WITH months AS (
			SELECT
				to_char(date_trunc('month', now()) - (n || ' month')::interval, 'YYYY-MM') AS year_month,
				(date_trunc('month', now()) - (n || ' month')::interval) AS start_date
			FROM generate_series(0, 11) AS n
		)
		SELECT
			m.year_month,
			COUNT(p.id) AS count
		FROM months m
		LEFT JOIN projects p
			ON to_char(p.created_at, 'YYYY-MM') = m.year_month
		LEFT JOIN project_members pm
			ON pm.project_id = p.id
			AND pm.user_id = $1
		GROUP BY m.year_month
		ORDER BY m.year_month
		`,
		mID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"project.ProjectRepository.ListCreatedByDay: %w",
			err,
		)
	}
	defer rows.Close()

	metrics := make([]project.Metric, 0)

	for rows.Next() {
		var (
			key   string
			value int64
		)

		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf(
				"project.ProjectRepository.ListCreatedByDay: %w",
				err,
			)
		}

		metrics = append(metrics, project.NewMetric(key, value))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"project.ProjectRepository.ListCreatedByDay: %w",
			err,
		)
	}

	return metrics, nil
}

func (r *ProjectRepository) Count(ctx context.Context, f project.Filter) (int, error) {
	ex := r.executor(ctx)

	var (
		sb     strings.Builder
		args   = make([]any, 0, 3)
		argPos = 1
	)

	sb.WriteString(`
		SELECT COUNT(DISTINCT p.id)
		FROM projects p
		INNER JOIN project_members pm
			ON p.id = pm.project_id
		WHERE pm.user_id = $`)
	sb.WriteString(strconv.Itoa(argPos))

	args = append(args, f.MemberID.String())
	argPos++

	if f.Keyword != nil {
		sb.WriteString(`
			AND p.name ILIKE $`)
		sb.WriteString(strconv.Itoa(argPos))

		args = append(args, "%"+*f.Keyword+"%")
		argPos++
	}

	if f.Status != nil {
		sb.WriteString(`
			AND p.status = $`)
		sb.WriteString(strconv.Itoa(argPos))

		args = append(args, f.Status.String())
		argPos++
	}

	var count int

	if err := ex.QueryRowContext(
		ctx,
		sb.String(),
		args...,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf(
			"project.ProjectRepository.Count: %w",
			err,
		)
	}

	return count, nil
}

func (r *ProjectRepository) Save(ctx context.Context, p *project.Project) error {
	ex := r.executor(ctx)

	res, err := ex.ExecContext(
		ctx,
		`UPDATE projects
		 SET name = $1,
		     status = $2,
		     version = version + 1,
		     updated_at = $3
		 WHERE id = $4 AND version = $5`,
		p.Name().String(),
		p.Status().String(),
		p.UpdatedAt(),
		p.ID().String(),
		p.Version(),
	)
	if err != nil {
		return fmt.Errorf("project.ProjectRepository.Save: update project: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("project.ProjectRepository.Save: rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf(
			"project.ProjectRepository.Save: %w",
			project.ErrProjectConcurrentModification,
		)
	}

	_, err = ex.ExecContext(
		ctx,
		`DELETE FROM project_members WHERE project_id = $1`,
		p.ID().String(),
	)
	if err != nil {
		return fmt.Errorf("project.ProjectRepository.Save: delete members: %w", err)
	}

	members := p.Members()
	if len(members) == 0 {
		return nil
	}

	var sb strings.Builder
	args := make([]any, 0, len(members)*2)

	sb.WriteString(`INSERT INTO project_members(project_id, user_id) VALUES `)

	for i := range members {
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

		args = append(args, p.ID().String(), members[i].String())
	}

	_, err = ex.ExecContext(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf("project.ProjectRepository.Save: insert members: %w", err)
	}

	return nil
}

func (r *ProjectRepository) SaveAll(ctx context.Context, projects []*project.Project) error {
	for _, p := range projects {
		if err := r.Save(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProjectRepository) Remove(ctx context.Context, p *project.Project) error {
	ex := r.executor(ctx)

	res, err := ex.ExecContext(
		ctx,
		`DELETE FROM projects
		 WHERE id = $1 AND version = $2`,
		p.ID().String(),
		p.Version(),
	)
	if err != nil {
		return fmt.Errorf("project.ProjectRepository.Remove: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("project.ProjectRepository.Remove: rows affected: %w", err)
	}

	if rows == 0 {
		return project.ErrProjectNotFound
	}

	return nil
}
