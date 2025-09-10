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

// GetInvoiceStatsByProject returns the total number of invoices/quotes
// found and a list of invoices with status for
// a specified project using the project ID.
// The search criteria parameter can be used to filter results
// by the type (invoice/quote), limit and offset results.
//
// If any error occurs, store.ErrQueryFailed is returned.
func (q *InvoiceStorePostgres) GetInvoiceStatsByProject(
	ctx context.Context,
	projectId int64,
	arg *params.InvoiceSearch) (*agg.WithCount[agg.InvoiceWithStatus], error) {

	var query strings.Builder
	var countQuery strings.Builder

	query.WriteString(`
	SELECT 
		i.id, 
		i.is_invoice, 
		i.issued_at, 
		i.due_at, 
		i.total, 
		i.currency_code, 
		ins.name
	FROM invoices i 
	JOIN invoice_statuses ins ON ins.id = i.status
	WHERE i.project_id = $1 
	`)

	countQuery.WriteString(`
	SELECT COUNT(i.id) FROM invoices i
	JOIN invoice_statuses ins ON ins.id = i.status
	WHERE i.project_id = $1 
	`)

	queryArgs := []any{projectId}
	countQueryArgs := []any{projectId}

	if arg.Type == params.InvoiceTypeInvoice {
		query.WriteString(" AND i.is_invoice = true")
		countQuery.WriteString(" AND i.is_invoice = true")
	}

	if arg.Type == params.InvoiceTypeQuote {
		query.WriteString(" AND i.is_invoice = false")
		countQuery.WriteString(" AND i.is_invoice = false")
	}

	paramCount := 1

	if len(arg.Status) > 0 {
		paramCount++
		query.WriteString(" AND ins.name = $")
		query.WriteString(strconv.Itoa(paramCount))
		countQuery.WriteString(" AND ins.name = $")
		countQuery.WriteString(strconv.Itoa(paramCount))

		queryArgs = append(queryArgs, arg.Status)
		countQueryArgs = append(countQueryArgs, arg.Status)
	}

	paramCount++
	query.WriteString(" LIMIT $")
	countQuery.WriteString(strconv.Itoa(paramCount))

	paramCount++
	query.WriteString(" OFFSET $")
	countQuery.WriteString(strconv.Itoa(paramCount))
	queryArgs = append(queryArgs, arg.Limit, arg.Offset)

	var totalInvoices int64
	err := q.db.QueryRowContext(ctx, countQuery.String(), countQueryArgs...).Scan(&totalInvoices)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	rows, err := q.db.QueryContext(ctx, query.String(), queryArgs...)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	result := agg.WithCount[agg.InvoiceWithStatus]{
		Total: totalInvoices,
		Items: make([]agg.InvoiceWithStatus, 0),
	}

	for rows.Next() {
		var row agg.InvoiceWithStatus
		err := rows.Scan(
			&row.Id,
			&row.IsInvoice,
			&row.IssuedAt,
			&row.DueAt,
			&row.Total,
			&row.CurrencyCode,
			&row.Status,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		result.Items = append(result.Items, row)
	}

	return &result, nil
}

// GetWithProjectByUserId returns the total number of invoices/quotes
// found and a list of invoices with project name for all projects
// that belong to the specified user.
// The search criteria parameter can be used to filter results
// by a keyword, the type (invoice/quote), limit and offset results.
//
// If any error occurs, store.ErrQueryFailed is returned.
func (q *InvoiceStorePostgres) GetWithProjectByUserId(
	ctx context.Context,
	userId int64,
	arg *params.InvoiceSearch) (*agg.WithCount[agg.InvoiceWithProject], error) {

	var query strings.Builder
	var countQuery strings.Builder

	query.WriteString(`
	SELECT 
		i.id,
		p.id AS project_id,
		p.name,
		i.is_invoice,
		i.issued_at,
		i.due_at,
		i.total,
		i.currency_code,
		ins.name
	FROM project_users pu
	JOIN projects p ON p.id = pu.project_id
	JOIN invoices i ON i.project_id = p.id
	JOIN invoice_statuses ins ON ins.id = i.status
	WHERE pu.user_id = $1
	`)

	countQuery.WriteString(`
	SELECT COUNT(i.id)
	FROM project_users pu
	JOIN projects p ON p.id = pu.project_id
	JOIN invoices i ON i.project_id = p.id
	JOIN invoice_statuses ins ON ins.id = i.status
	WHERE pu.user_id = $1
	`)

	queryArgs := []any{userId}
	countQueryArgs := []any{userId}

	paramCount := 1

	if arg.Type == params.InvoiceTypeInvoice {
		query.WriteString(" AND i.is_invoice = true")
		countQuery.WriteString(" AND i.is_invoice = true")
	}

	if arg.Type == params.InvoiceTypeQuote {
		query.WriteString(" AND i.is_invoice = false")
		countQuery.WriteString(" AND i.is_invoice = false")
	}

	if len(arg.Keyword) > 0 {
		paramCount++

		query.WriteString(" AND p.name LIKE $")
		query.WriteString(strconv.Itoa(paramCount))

		countQuery.WriteString(" AND p.name LIKE $")
		countQuery.WriteString(strconv.Itoa(paramCount))

		keyword := "%" + arg.Keyword + "%"

		queryArgs = append(queryArgs, keyword)
		countQueryArgs = append(countQueryArgs, keyword)
	}

	if len(arg.Status) > 0 {
		paramCount++

		query.WriteString(" AND ins.name = $")
		query.WriteString(strconv.Itoa(paramCount))

		countQuery.WriteString(" AND ins.name = $")
		countQuery.WriteString(strconv.Itoa(paramCount))

		queryArgs = append(queryArgs, arg.Status)
		countQueryArgs = append(countQueryArgs, arg.Status)
	}

	paramCount++
	query.WriteString(" LIMIT $")
	query.WriteString(strconv.Itoa(paramCount))

	paramCount++
	query.WriteString(" OFFSET $")

	query.WriteString(strconv.Itoa(paramCount))
	queryArgs = append(queryArgs, arg.Limit, arg.Offset)

	var totalInvoices int64
	err := q.db.QueryRowContext(ctx, countQuery.String(), countQueryArgs...).Scan(&totalInvoices)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	rows, err := q.db.QueryContext(ctx, query.String(), queryArgs...)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	result := agg.WithCount[agg.InvoiceWithProject]{
		Total: totalInvoices,
		Items: make([]agg.InvoiceWithProject, 0),
	}

	for rows.Next() {
		var row agg.InvoiceWithProject
		err := rows.Scan(
			&row.Id,
			&row.ProjectId,
			&row.ProjectName,
			&row.IsInvoice,
			&row.IssuedAt,
			&row.DueAt,
			&row.Total,
			&row.CurrencyCode,
			&row.Status,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		result.Items = append(result.Items, row)
	}

	return &result, nil
}

// GetWithDetailsById returns a detailed invoice with invoice items.
// The invoice is specified by the invoice ID.
// This function returns a pointer to an agg.InvoiceDetails.
//
// If any error occurs, store.ErrQueryFailed is returned.
func (q *InvoiceStorePostgres) GetWithDetailsById(
	ctx context.Context,
	invoiceId int64) (*agg.InvoiceDetails, error) {

	var invoiceQuery strings.Builder
	var itemsQuery strings.Builder

	invoiceQuery.WriteString(`
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
	`)

	itemsQuery.WriteString(`
	SELECT 
		id,
		description,
		qty,
		unit_price,
		unit_discount,
		discount_type,
		unit_tax,
		tax_type,
		tax,
		discount,
		total
	FROM invoice_items
	WHERE invoice_id = $1
	`)

	var invoice agg.InvoiceDetails
	invoice.Items = make([]agg.InvoiceItem, 0)

	err := q.db.QueryRowContext(ctx, invoiceQuery.String(), invoiceId).Scan(
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

	rows, err := q.db.QueryContext(ctx, itemsQuery.String(), invoiceId)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

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
			&row.Tax,
			&row.Discount,
			&row.Total,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		invoice.Items = append(invoice.Items, row)
	}

	return &invoice, nil
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
		return false, store.ErrInsertFailed
	}

	if err = tx.Commit(); err != nil {
		return false, store.ErrInsertFailed
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
func (q *InvoiceStorePostgres) CancelById(ctx context.Context, invoiceId int64) (bool, error) {

	query := `
	SELECT
  		ins.name = $1 AS pending
	FROM invoices i
	JOIN invoice_statuses ins
  	ON ins.id = i.status
	WHERE i.id = $2
	`

	var isPending bool
	err := q.db.QueryRowContext(ctx, query, params.InvoiceStatusPending, invoiceId).Scan(&isPending)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	if !isPending {
		return true, nil
	}

	setStatusQuery := `
	UPDATE invoices
	SET status = (SELECT id from invoice_statuses WHERE name = $1)
	WHERE id = $2
	`

	_, err = q.db.ExecContext(ctx, setStatusQuery, params.InvoiceStatusCancelled, invoiceId)
	if err != nil {
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
func (q *InvoiceStorePostgres) PayById(ctx context.Context, invoiceId int64) (bool, error) {

	query := `
	SELECT i.is_invoice, ins.name AS status
	FROM invoices i 
	JOIN invoice_statuses ins ON ins.id = i.status
	WHERE i.id = $1
	`

	var isInvoice bool
	var invStatus string

	err := q.db.QueryRowContext(ctx, query, invoiceId).Scan(&isInvoice, &invStatus)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	if !isInvoice {
		return false, store.ErrUnexpectedType
	}

	if invStatus != params.InvoiceStatusAccepted {
		return true, nil
	}

	setStatusQuery := `
	UPDATE invoices
	SET status = (SELECT id FROM invoice_statuses WHERE name = $1)
	WHERE id = $2
	`

	_, err = q.db.ExecContext(ctx, setStatusQuery, params.InvoiceStatusPaid, invoiceId)
	if err != nil {
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
func (q *InvoiceStorePostgres) QuoteToInvoice(ctx context.Context, quoteId int64) (bool, error) {

	query := `
	SELECT i.is_invoice, ins.name AS status
	FROM invoices i 
	JOIN invoice_statuses ins ON ins.id = i.status
	WHERE i.id = ?
	`

	var isInvoice bool
	var invStatus string

	err := q.db.QueryRowContext(ctx, query, quoteId).Scan(&isInvoice, &invStatus)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	if isInvoice {
		return true, nil
	}

	if invStatus != params.InvoiceStatusPending {
		return false, store.ErrUnexpectedType
	}

	setInvoiceQuery := "UPDATE invoices SET is_invoice = true WHERE id = ?"

	_, err = q.db.ExecContext(ctx, setInvoiceQuery, quoteId)
	if err != nil {
		return false, store.ErrUpdateFailed
	}

	return false, nil

}
