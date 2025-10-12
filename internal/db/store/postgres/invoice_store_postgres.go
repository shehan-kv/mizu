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

// Postgres implementation of InvoiceStore interface
type InvoiceStorePostgres struct {
	db *sql.DB
}

// Creates a new instance of an InvoiceStorePostgres
//
// Parameters:
//   - db: *sql.DB
//
// Returns:
//   - *InvoiceStorePostgres
func NewInvoiceStore(db *sql.DB) *InvoiceStorePostgres {
	return &InvoiceStorePostgres{db: db}
}

// Implementing CreateOne defined in InvoiceStore interface
func (q *InvoiceStorePostgres) CreateOne(ctx context.Context, userId int64, arg *params.InvoiceCreate) (int64, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, store.ErrInsertFailed
	}

	defer tx.Rollback()

	insertInvoice := `
	INSERT INTO invoices(project_id, is_invoice, status, total, discount, tax, currency_code, note)
	VALUES($1, $2, (SELECT id FROM invoice_statuses WHERE name = $3), $4, $5, $6, $7, $8) RETURNING id
	`

	var invoiceId int64
	err = tx.QueryRowContext(ctx, insertInvoice,
		arg.ProjectId,
		arg.IsInvoice,
		arg.Status,
		arg.Total,
		arg.Discount,
		arg.Tax,
		arg.CurrencyCode,
		arg.Note).Scan(&invoiceId)

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "foreign_key_violation" {
				return 0, store.ErrForeignKeyViolation
			}

			if err.Code.Name() == "not_null_violation" {
				return 0, store.ErrNotNullViolation
			}
		}

		return 0, store.ErrInsertFailed
	}

	insertItem := `
	INSERT INTO invoice_items(
		invoice_id, description, qty, unit_price, tax, tax_type, discount, discount_type, total)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	for _, item := range arg.Items {
		if _, err = tx.ExecContext(ctx, insertItem,
			invoiceId,
			item.Description,
			item.Qty,
			item.UnitPrice,
			item.Tax,
			item.TaxType,
			item.Discount,
			item.DiscountType,
			item.Total); err != nil {

			if err, ok := err.(*pq.Error); ok {
				if err.Code.Name() == "not_null_violation" {
					return 0, store.ErrNotNullViolation
				}

				if err.Code.Name() == "check_violation" {
					return 0, store.ErrCheckViolation
				}
			}
			return 0, store.ErrInsertFailed
		}
	}

	insertInvHist := `
	INSERT INTO invoice_history (invoice_id, user_id, is_invoice, event, new_status)
	VALUES ( 
		$1, $2, $3,
		(SELECT id FROM invoice_history_events WHERE name = $4),
		(SELECT id FROM invoice_statuses WHERE name = $5)
	)
	`

	_, err = tx.ExecContext(
		ctx,
		insertInvHist,
		invoiceId,
		userId,
		arg.IsInvoice,
		params.InvoiceHistoryEventCreated,
		arg.Status,
	)

	if err != nil {
		return 0, store.ErrInsertFailed
	}

	if err = tx.Commit(); err != nil {
		return 0, store.ErrInsertFailed
	}

	return invoiceId, nil
}

// GetSummaryByProjectId returns the total number of invoices/quotes
// found and a list of invoices for
// a specified project using the project ID.
// The search criteria parameter can be used to filter results
// by the type (invoice/quote), limit and offset results.
//
// If any error occurs, store.ErrQueryFailed is returned.
func (q *InvoiceStorePostgres) GetSummaryByProjectId(
	ctx context.Context,
	projectId int64,
	arg *params.InvoiceSearch) ([]agg.Invoice, error) {

	var query strings.Builder

	query.WriteString(`
	SELECT 
		i.id,
		p.id,
		p.name,
		i.is_invoice,
		ins.name,
		i.issued_at,
		i.due_at,
		i.total,
		i.discount,
		i.tax,
		i.currency_code,
		i.note
	FROM invoices i
	JOIN invoice_statuses ins ON ins.id = i.status
	JOIN projects p ON p.id = i.project_id
	WHERE p.id = $1
	`)

	queryArgs := []any{projectId}

	switch arg.Type {
	case params.InvoiceTypeInvoice:
		query.WriteString(" AND i.is_invoice = true")

	case params.InvoiceTypeQuote:
		query.WriteString(" AND i.is_invoice = false")
	}

	paramCount := 1

	if len(arg.Status) > 0 {
		paramCount++
		query.WriteString(" AND ins.name = $")
		query.WriteString(strconv.Itoa(paramCount))

		queryArgs = append(queryArgs, arg.Status)
	}

	paramCount++
	query.WriteString(" LIMIT $")
	query.WriteString(strconv.Itoa(paramCount))

	paramCount++
	query.WriteString(" OFFSET $")
	query.WriteString(strconv.Itoa(paramCount))
	queryArgs = append(queryArgs, arg.Limit, arg.Offset)

	rows, err := q.db.QueryContext(ctx, query.String(), queryArgs...)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	result := make([]agg.Invoice, 0)

	for rows.Next() {
		var row agg.Invoice
		err := rows.Scan(
			&row.Id,
			&row.ProjectId,
			&row.ProjectName,
			&row.IsInvoice,
			&row.Status,
			&row.IssuedAt,
			&row.DueAt,
			&row.Total,
			&row.Discount,
			&row.Tax,
			&row.CurrencyCode,
			&row.Note,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		result = append(result, row)
	}

	return result, nil
}

func (q *InvoiceStorePostgres) CountSummaryByProjectId(
	ctx context.Context,
	projectId int64,
	arg *params.InvoiceSearch) (int64, error) {

	var query strings.Builder

	query.WriteString(`
	SELECT COUNT(i.id) FROM invoices i
	JOIN invoice_statuses ins ON ins.id = i.status
	WHERE i.project_id = $1 
	`)

	queryArgs := []any{projectId}

	switch arg.Type {
	case params.InvoiceTypeInvoice:
		query.WriteString(" AND i.is_invoice = true")

	case params.InvoiceTypeQuote:
		query.WriteString(" AND i.is_invoice = false")
	}

	paramCount := 1

	if len(arg.Status) > 0 {
		paramCount++
		query.WriteString(" AND ins.name = $")
		query.WriteString(strconv.Itoa(paramCount))

		queryArgs = append(queryArgs, arg.Status)
	}

	var count int64
	err := q.db.QueryRowContext(ctx, query.String(), queryArgs...).Scan(&count)
	if err != nil {
		return 0, store.ErrQueryFailed
	}

	return count, nil
}

// GetSummaryByUserId returns the total number of invoices/quotes
// found and a list of invoices with project name for all projects
// that belong to the specified user.
// The search criteria parameter can be used to filter results
// by a keyword, the type (invoice/quote), limit and offset results.
//
// If any error occurs, store.ErrQueryFailed is returned.
func (q *InvoiceStorePostgres) GetSummaryByUserId(
	ctx context.Context,
	userId int64,
	arg *params.InvoiceSearch) ([]agg.Invoice, error) {

	var query strings.Builder
	query.WriteString(`
	SELECT 
		i.id,
		p.id AS project_id,
		p.name,
		i.is_invoice,
		ins.name,
		i.issued_at,
		i.due_at,
		i.total,
		i.discount,
		i.tax,
		i.currency_code,
		i.note
	FROM project_users pu
	JOIN projects p ON p.id = pu.project_id
	JOIN invoices i ON i.project_id = p.id
	JOIN invoice_statuses ins ON ins.id = i.status
	WHERE pu.user_id = $1
	`)

	queryArgs := []any{userId}

	switch arg.Type {
	case params.InvoiceTypeInvoice:
		query.WriteString(" AND i.is_invoice = true")

	case params.InvoiceTypeQuote:
		query.WriteString(" AND i.is_invoice = false")
	}

	paramCount := 1
	if len(arg.Keyword) > 0 {
		paramCount++

		query.WriteString(" AND p.name LIKE $")
		query.WriteString(strconv.Itoa(paramCount))

		queryArgs = append(queryArgs, "%"+arg.Keyword+"%")
	}

	if len(arg.Status) > 0 {
		paramCount++

		query.WriteString(" AND ins.name = $")
		query.WriteString(strconv.Itoa(paramCount))

		queryArgs = append(queryArgs, arg.Status)
	}

	paramCount++
	query.WriteString(" LIMIT $")
	query.WriteString(strconv.Itoa(paramCount))

	paramCount++
	query.WriteString(" OFFSET $")

	query.WriteString(strconv.Itoa(paramCount))
	queryArgs = append(queryArgs, arg.Limit, arg.Offset)

	rows, err := q.db.QueryContext(ctx, query.String(), queryArgs...)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	result := make([]agg.Invoice, 0)

	for rows.Next() {
		var row agg.Invoice
		err := rows.Scan(
			&row.Id,
			&row.ProjectId,
			&row.ProjectName,
			&row.IsInvoice,
			&row.Status,
			&row.IssuedAt,
			&row.DueAt,
			&row.Total,
			&row.Discount,
			&row.Tax,
			&row.CurrencyCode,
			&row.Note,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		result = append(result, row)
	}

	return result, nil
}

func (q *InvoiceStorePostgres) CountSummaryByUserId(
	ctx context.Context,
	userId int64,
	arg *params.InvoiceSearch) (int64, error) {

	var query strings.Builder
	query.WriteString(`
	SELECT COUNT(i.id)
	FROM project_users pu
	JOIN projects p ON p.id = pu.project_id
	JOIN invoices i ON i.project_id = pu.project_id
	JOIN invoice_statuses ins ON ins.id = i.status
	WHERE pu.user_id = $1
	`)

	queryArgs := []any{userId}

	switch arg.Type {
	case params.InvoiceTypeInvoice:
		query.WriteString(" AND i.is_invoice = true")

	case params.InvoiceTypeQuote:
		query.WriteString(" AND i.is_invoice = false")
	}

	paramCount := 1
	if len(arg.Keyword) > 0 {
		paramCount++

		query.WriteString(" AND p.name LIKE $")
		query.WriteString(strconv.Itoa(paramCount))

		queryArgs = append(queryArgs, "%"+arg.Keyword+"%")
	}

	if len(arg.Status) > 0 {
		paramCount++

		query.WriteString(" AND ins.name = $")
		query.WriteString(strconv.Itoa(paramCount))

		queryArgs = append(queryArgs, arg.Status)
	}

	var count int64
	err := q.db.QueryRowContext(ctx, query.String(), queryArgs...).Scan(&count)
	if err != nil {
		return 0, store.ErrQueryFailed
	}

	return count, nil
}

func (q *InvoiceStorePostgres) GetSummaryById(ctx context.Context, invoiceId int64) (*agg.Invoice, error) {

	query := `
	SELECT 
		i.id,
		p.id,
		p.name,
		i.is_invoice,
		ins.name,
		i.issued_at,
		i.due_at,
		i.total,
		i.discount,
		i.tax,
		i.currency_code,
		i.note
	FROM invoices i
	JOIN invoice_statuses ins ON ins.id = i.status
	JOIN projects p ON p.id = i.project_id
	WHERE i.id = $1
	`

	var invoice agg.Invoice
	err := q.db.QueryRowContext(ctx, query, invoiceId).Scan(
		&invoice.Id,
		&invoice.ProjectId,
		&invoice.ProjectName,
		&invoice.IsInvoice,
		&invoice.Status,
		&invoice.IssuedAt,
		&invoice.DueAt,
		&invoice.Total,
		&invoice.Discount,
		&invoice.Tax,
		&invoice.CurrencyCode,
		&invoice.Note,
	)

	if err != nil {
		return nil, store.ErrQueryFailed
	}
	return &invoice, nil
}

func (q *InvoiceStorePostgres) GetItemsByInvoiceId(
	ctx context.Context,
	invoiceId int64) ([]agg.InvoiceItem, error) {

	query := `
	SELECT 
		id,
		description,
		qty,
		unit_price,
		unit_discount,
		discount_type,
		unit_tax,
		tax_type,
		total_tax,
		total_discount,
		total
	FROM invoice_items
	WHERE invoice_id = $1
	`

	rows, err := q.db.QueryContext(ctx, query, invoiceId)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	result := make([]agg.InvoiceItem, 0)

	for rows.Next() {
		var row agg.InvoiceItem
		err := rows.Scan(
			&row.Id,
			&row.Description,
			&row.Qty,
			&row.UnitPrice,
			&row.UnitDiscount,
			&row.DiscountType,
			&row.UnitTax,
			&row.TaxType,
			&row.TotalTax,
			&row.TotalDiscount,
			&row.Total,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		result = append(result, row)
	}

	return result, nil
}

func (q *InvoiceStorePostgres) GetHistoryByInvoiceId(
	ctx context.Context,
	invoiceId int64) ([]agg.InvoiceHistory, error) {

	query := `
	SELECT
		ih.id,
		ih.user_id,
		u.first_name,
		u.last_name,
		u.image,
		u.title,
		r.name AS role,
		ihe.name AS event,
		ih.recorded_at,
		ih.is_invoice,
		ins1.name AS last_status,
		ins2.name AS new_status
	FROM invoice_history ih
	JOIN users u ON u.id = ih.user_id
	JOIN roles r ON r.id = u.role
	JOIN invoice_history_events ihe ON ihe.id = ih.event
	LEFT JOIN invoice_statuses ins1 ON ins1.id = ih.last_status
	LEFT JOIN invoice_statuses ins2 ON ins2.id = ih.new_status
	WHERE ih.invoice_id = $1 
	`

	result := make([]agg.InvoiceHistory, 0)

	rows, err := q.db.QueryContext(ctx, query, invoiceId)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	for rows.Next() {
		var row agg.InvoiceHistory
		err := rows.Scan(
			&row.Id,
			&row.UserId,
			&row.FirstName,
			&row.LastName,
			&row.Image,
			&row.Title,
			&row.Role,
			&row.Event,
			&row.RecordedAt,
			&row.IsInvoice,
			&row.LastStatus,
			&row.NewStatus,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		result = append(result, row)
	}

	return result, nil
}

// AcceptById marks a specified invoice as accepted.
// The invoice is specified by invoice ID.
// Returns a boolean and an error.
//
// If the returned boolean is:
//   - true: the invoice is already accepted
//   - false: successfully accepted the invoice
//
// If any error occurs, store.ErrUpdateFailed is returned.
func (q *InvoiceStorePostgres) AcceptById(ctx context.Context, userId int64, invoiceId int64) (bool, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	defer tx.Rollback()

	query := `
	SELECT
  		ins.name = $1 AS pending
	FROM invoices i
	JOIN invoice_statuses ins
  	ON ins.id = i.status
	WHERE i.id = $2
	`

	var isPending bool
	err = tx.QueryRowContext(ctx, query, params.InvoiceStatusPending, invoiceId).Scan(&isPending)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	if !isPending {
		return true, nil
	}

	setStatusQuery := `
	UPDATE invoices
	SET is_invoice = true, status = (SELECT id from invoice_statuses WHERE name = $1)
	WHERE id = $2
	`

	_, err = tx.ExecContext(ctx, setStatusQuery, params.InvoiceStatusAccepted, invoiceId)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	insertInvHist := `
	INSERT INTO invoice_history (invoice_id, user_id, is_invoice, event, last_status, new_status)
	VALUES ( 
		$1, $2, $3,
		(SELECT id FROM invoice_history_events WHERE name = $4),
		(SELECT id FROM invoice_statuses WHERE name = $5),
		(SELECT id FROM invoice_statuses WHERE name = $6)
	)
	`

	_, err = tx.ExecContext(
		ctx,
		insertInvHist,
		invoiceId,
		userId,
		true,
		params.InvoiceHistoryEventAccepted,
		params.InvoiceStatusPending,
		params.InvoiceStatusAccepted,
	)

	if err != nil {
		return false, store.ErrUpdateFailed
	}

	if err = tx.Commit(); err != nil {
		return false, store.ErrUpdateFailed
	}

	return false, nil
}

// RejectById marks a specified invoice as rejected.
// The invoice is specified by invoice ID.
// Returns a boolean and an error.
//
// If the returned boolean is:
//   - true: the invoice is already rejected
//   - false: successfully rejected the invoice
//
// If any error occurs, store.ErrUpdateFailed is returned.
func (q *InvoiceStorePostgres) RejectById(ctx context.Context, userId int64, invoiceId int64) (bool, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	defer tx.Rollback()

	query := `
	SELECT
  		ins.name AS status,
		i.is_invoice
	FROM invoices i
	JOIN invoice_statuses ins
  	ON ins.id = i.status
	WHERE i.id = $1
	`

	var status string
	var isInvoice bool
	err = tx.QueryRowContext(ctx, query, invoiceId).Scan(&status, &isInvoice)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	if status != params.InvoiceStatusPending {
		return true, nil
	}

	setStatusQuery := `
	UPDATE invoices
	SET status = (SELECT id from invoice_statuses WHERE name = $1)
	WHERE id = $2
	`

	_, err = tx.ExecContext(ctx, setStatusQuery, params.InvoiceStatusRejected, invoiceId)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	insertInvHist := `
	INSERT INTO invoice_history (invoice_id, user_id, is_invoice, event, last_status, new_status)
	VALUES ( 
		$1, $2, $3,
		(SELECT id FROM invoice_history_events WHERE name = $4),
		(SELECT id FROM invoice_statuses WHERE name = $5),
		(SELECT id FROM invoice_statuses WHERE name = $6)
	)
	`

	_, err = tx.ExecContext(
		ctx,
		insertInvHist,
		invoiceId,
		userId,
		isInvoice,
		params.InvoiceHistoryEventRejected,
		status,
		params.InvoiceStatusRejected,
	)

	if err != nil {
		return false, store.ErrUpdateFailed
	}

	if err = tx.Commit(); err != nil {
		return false, store.ErrUpdateFailed
	}

	return false, nil
}

// CancelById marks a specified invoice or a quote as cancelled.
// The invoice/quote is specified by invoice ID.
// Returns a boolean and an error.
//
// If the returned boolean is:
//   - true: the invoice is already cancelled
//   - false: successfully cancelled the invoice
//
// If any error occurs, store.ErrUpdateFailed is returned.
func (q *InvoiceStorePostgres) CancelById(ctx context.Context, userId int64, invoiceId int64) (bool, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	defer tx.Rollback()

	query := `
	SELECT
  		ins.name AS status,
		i.is_invoice
	FROM invoices i
	JOIN invoice_statuses ins
  	ON ins.id = i.status
	WHERE i.id = $1
	`

	var status string
	var isInvoice bool
	err = tx.QueryRowContext(ctx, query, invoiceId).Scan(&status, &isInvoice)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	if status != params.InvoiceStatusPending {
		return true, nil
	}

	setStatusQuery := `
	UPDATE invoices
	SET status = (SELECT id from invoice_statuses WHERE name = $1)
	WHERE id = $2
	`

	_, err = tx.ExecContext(ctx, setStatusQuery, params.InvoiceStatusCancelled, invoiceId)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	insertInvHist := `
	INSERT INTO invoice_history (invoice_id, user_id, is_invoice, event, last_status, new_status)
	VALUES ( 
		$1, $2, $3,
		(SELECT id FROM invoice_history_events WHERE name = $4),
		(SELECT id FROM invoice_statuses WHERE name = $5),
		(SELECT id FROM invoice_statuses WHERE name = $6)
	)
	`

	_, err = tx.ExecContext(
		ctx,
		insertInvHist,
		invoiceId,
		userId,
		isInvoice,
		params.InvoiceHistoryEventCancelled,
		status,
		params.InvoiceStatusRejected,
	)

	if err != nil {
		return false, store.ErrUpdateFailed
	}

	if err = tx.Commit(); err != nil {
		return false, store.ErrUpdateFailed
	}

	return false, nil
}

// PayById marks a specified invoice as paid.
// The invoice is specified by invoice ID.
// Returns a boolean and an error.
//
// If the returned boolean is:
//   - true: the invoice is already paid
//   - false: successfully marked the invoice as paid
//
// Errors:
//   - if the provided ID points to a quote, store.ErrUnexpectedType is returned.
//   - If any other error occurs, store.ErrUpdateFailed is returned.
func (q *InvoiceStorePostgres) PayById(ctx context.Context, userId int64, invoiceId int64) (bool, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	defer tx.Rollback()

	query := `
	SELECT i.is_invoice, ins.name AS status
	FROM invoices i 
	JOIN invoice_statuses ins ON ins.id = i.status
	WHERE i.id = $1
	`

	var isInvoice bool
	var status string

	err = tx.QueryRowContext(ctx, query, invoiceId).Scan(&isInvoice, &status)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	if !isInvoice {
		return false, store.ErrUnexpectedType
	}

	if status != params.InvoiceStatusAccepted {
		return true, nil
	}

	setStatusQuery := `
	UPDATE invoices
	SET status = (SELECT id FROM invoice_statuses WHERE name = $1)
	WHERE id = $2
	`

	_, err = tx.ExecContext(ctx, setStatusQuery, params.InvoiceStatusPaid, invoiceId)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	insertInvHist := `
	INSERT INTO invoice_history (invoice_id, user_id, is_invoice, event, last_status, new_status)
	VALUES ( 
		$1, $2, $3,
		(SELECT id FROM invoice_history_events WHERE name = $4),
		(SELECT id FROM invoice_statuses WHERE name = $5),
		(SELECT id FROM invoice_statuses WHERE name = $6)
	)
	`

	_, err = tx.ExecContext(
		ctx,
		insertInvHist,
		invoiceId,
		userId,
		isInvoice,
		params.InvoiceHistoryEventPaid,
		status,
		params.InvoiceStatusPaid,
	)

	if err != nil {
		return false, store.ErrUpdateFailed
	}

	if err = tx.Commit(); err != nil {
		return false, store.ErrUpdateFailed
	}

	return false, nil
}

// QuoteToInvoice converts a quote to an invoice.
// The quote is specified by quote ID.
// Returns a boolean and an error.
//
// If the returned boolean is:
//   - true: the quote is already converted
//   - false: successfully converted to an invoice
//
// Errors:
//   - if the quote status is invalid, store.ErrUnexpectedType is returned.
//   - If any other error occurs, store.ErrUpdateFailed is returned.
func (q *InvoiceStorePostgres) QuoteToInvoice(ctx context.Context, userId int64, quoteId int64) (bool, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	defer tx.Rollback()

	query := `
	SELECT i.is_invoice, ins.name AS status
	FROM invoices i 
	JOIN invoice_statuses ins ON ins.id = i.status
	WHERE i.id = $1
	`

	var isInvoice bool
	var status string

	err = tx.QueryRowContext(ctx, query, quoteId).Scan(&isInvoice, &status)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	if isInvoice {
		return true, nil
	}

	if status != params.InvoiceStatusPending {
		return false, store.ErrUnexpectedType
	}

	setInvoiceQuery := "UPDATE invoices SET is_invoice = true WHERE id = $1"

	_, err = tx.ExecContext(ctx, setInvoiceQuery, quoteId)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	insertInvHist := `
	INSERT INTO invoice_history (invoice_id, user_id, is_invoice, event, last_status, new_status)
	VALUES ( 
		$1, $2, $3,
		(SELECT id FROM invoice_history_events WHERE name = $4),
		(SELECT id FROM invoice_statuses WHERE name = $5),
		(SELECT id FROM invoice_statuses WHERE name = $6)
	)
	`

	_, err = tx.ExecContext(
		ctx,
		insertInvHist,
		quoteId,
		userId,
		true,
		params.InvoiceHistoryEventConverted,
		status,
		status,
	)

	if err != nil {
		return false, store.ErrUpdateFailed
	}

	if err = tx.Commit(); err != nil {
		return false, store.ErrUpdateFailed
	}

	return false, nil
}
