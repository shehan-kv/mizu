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

// ChangeRequestStorePostgres implements the ChangeRequestStore interface using Postgres.
// It persists Contract entities in a SQLite database via the provided *sql.DB.
// The db connection must be non-nil, open, and initialized with the required
// contracts schema. Methods wrap lower-level SQL errors into domain-specific
// errors defined in the store package.
type ChangeRequestStorePostgres struct {
	db *sql.DB
}

// NewChangeRequestStore constructs a ChangeRequestStorePostgres that persists
// ChangeRequest entities in a SQLite database via the provided *sql.DB.
func NewChangeRequestStore(db *sql.DB) *ChangeRequestStorePostgres {
	return &ChangeRequestStorePostgres{db: db}
}

func (q *ChangeRequestStorePostgres) CreateOne(
	ctx context.Context,
	userId int64,
	arg *params.ChangeRequestCreate) (int64, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, store.ErrInsertFailed
	}

	defer tx.Rollback()

	reqQuery := `
	INSERT INTO change_requests(project_id, req_user_id, title, status)
	VALUES($1, $2, $3, (SELECT id FROM change_request_statuses WHERE name = $4)) RETURNING id
	`

	var reqId int64
	err = tx.QueryRowContext(
		ctx,
		reqQuery,
		arg.ProjectId,
		userId,
		arg.Title,
		params.ChangeRequestInProgress,
	).Scan(&reqId)

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "foreign_key_violation" {
				return 0, store.ErrForeignKeyViolation
			}

			if err.Code.Name() == "not_null_violation" {
				return 0, store.ErrNotNullViolation
			}

			return 0, store.ErrInsertFailed
		}
	}

	reqEntryQuery := `
	INSERT INTO change_request_entries(request_id, user_id, content)
	VALUES($1, $2, $3)
	`

	_, err = tx.ExecContext(ctx, reqEntryQuery, reqId, userId, arg.Content)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "foreign_key_violation" {
				return 0, store.ErrForeignKeyViolation
			}

			if err.Code.Name() == "not_null_violation" {
				return 0, store.ErrNotNullViolation
			}

			return 0, store.ErrInsertFailed
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, store.ErrInsertFailed
	}

	return reqId, nil
}

func (q *ChangeRequestStorePostgres) CreateEntry(
	ctx context.Context,
	userId int64,
	requestId int64,
	content string) error {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return store.ErrInsertFailed
	}

	defer tx.Rollback()

	reqQuery := `
	SELECT crs.name AS status
	FROM change_requests cr
	JOIN change_request_statuses crs ON crs.id = cr.status
	WHERE cr.id = $1
	`

	var reqStatus string
	err = tx.QueryRowContext(ctx, reqQuery, requestId).Scan(&reqStatus)
	if err != nil {
		return store.ErrInsertFailed
	}

	if reqStatus == params.ChangeRequestClosed {
		return store.ErrUnexpectedType
	}
	setStatusQuery := `
	UPDATE change_requests
	SET status = (SELECT id FROM change_request_statuses WHERE name = $1)
	WHERE id = $2
	`

	_, err = tx.ExecContext(ctx, setStatusQuery, params.ChangeRequestWaiting, requestId)
	if err != nil {
		return store.ErrInsertFailed
	}

	entryQuery := `
	INSERT INTO change_request_entries(request_id, user_id, content)
	VALUES ($1, $2, $3)
	`

	_, err = tx.ExecContext(ctx, entryQuery, requestId, userId, content)

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "foreign_key_violation" {
				return store.ErrForeignKeyViolation
			}

			if err.Code.Name() == "not_null_violation" {
				return store.ErrNotNullViolation
			}

			return store.ErrInsertFailed
		}
	}

	if err = tx.Commit(); err != nil {
		return store.ErrInsertFailed
	}

	return nil
}

func (q *ChangeRequestStorePostgres) GetByProjectId(
	ctx context.Context,
	projectId int64,
	arg *params.ChangeRequestSearch) (*agg.WithCount[agg.ChangeRequest], error) {

	var reqQuery strings.Builder

	reqQuery.WriteString(`
	SELECT 
		cr.id,
		cr.title,
		cr.created_at,
		cr.req_user_id AS user_id,
		u.first_name,
		u.last_name,
		p.id AS project_id,
		p.name AS project_name,
		crs.name AS status
	FROM change_requests cr
	JOIN projects p ON p.id = cr.project_id
	JOIN users u ON u.id = cr.req_user_id
	JOIN change_request_statuses crs ON crs.id = cr.status
	WHERE p.id = $1
	`)

	var countQuery strings.Builder
	countQuery.WriteString(`
	SELECT COUNT(cr.id)
	FROM change_requests cr
	JOIN projects p ON p.id = cr.project_id
	JOIN users u ON u.id = cr.req_user_id
	JOIN change_request_statuses crs ON crs.id = cr.status
	WHERE p.id = $1
	`)

	reqQueryArgs := []any{projectId}
	countQueryArgs := []any{projectId}

	paramCount := 1

	if len(arg.Keyword) != 0 {
		paramCount++
		reqQuery.WriteString(" AND cr.title = $")
		reqQuery.WriteString(strconv.Itoa(paramCount))
		countQuery.WriteString(" AND cr.title = $")
		countQuery.WriteString(strconv.Itoa(paramCount))

		reqQueryArgs = append(reqQueryArgs, arg.Keyword)
		countQueryArgs = append(countQueryArgs, arg.Keyword)
	}

	if len(arg.Status) != 0 {
		paramCount++
		reqQuery.WriteString(" AND crs.name = $")
		reqQuery.WriteString(strconv.Itoa(paramCount))
		countQuery.WriteString(" AND crs.name = $")
		countQuery.WriteString(strconv.Itoa(paramCount))

		reqQueryArgs = append(reqQueryArgs, arg.Status)
		countQueryArgs = append(countQueryArgs, arg.Status)
	}

	reqQuery.WriteString(" ORDER BY cr.created_at DESC")

	paramCount++
	reqQuery.WriteString(" LIMIT $")
	reqQuery.WriteString(strconv.Itoa(paramCount))

	paramCount++
	reqQuery.WriteString(" OFFSET $")
	reqQuery.WriteString(strconv.Itoa(paramCount))
	reqQueryArgs = append(reqQueryArgs, arg.Limit, arg.Offset)

	var totalRequests int64
	err := q.db.QueryRowContext(ctx, countQuery.String(), countQueryArgs...).Scan(&totalRequests)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	reqRows, err := q.db.QueryContext(ctx, reqQuery.String(), reqQueryArgs...)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer reqRows.Close()

	result := agg.WithCount[agg.ChangeRequest]{
		Total: totalRequests,
		Items: make([]agg.ChangeRequest, 0),
	}

	for reqRows.Next() {
		var row agg.ChangeRequest
		err := reqRows.Scan(
			&row.Id,
			&row.Title,
			&row.CreatedAt,
			&row.ReqUserId,
			&row.ReqUserFirstName,
			&row.ReqUserLastName,
			&row.ProjectId,
			&row.ProjectName,
			&row.Status,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		result.Items = append(result.Items, row)
	}

	return &result, nil
}
