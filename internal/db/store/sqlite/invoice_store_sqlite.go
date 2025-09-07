package sqlite

import (
	"context"
	"database/sql"
	agg "mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"strings"

	"github.com/mattn/go-sqlite3"
)

// SQLite implementation of InvoiceStore interface
type InvoiceStoreSqlite struct {
	db *sql.DB
}

// Creates a new instance of an InvoiceStoreSqlite
//
// Parameters:
//   - db: *sql.DB
//
// Returns:
//   - *InvoiceStoreSqlite
func NewInvoiceStore(db *sql.DB) *InvoiceStoreSqlite {
	return &InvoiceStoreSqlite{db: db}
}

// Implementing CreateOne defined in InvoiceStore interface
func (q *InvoiceStoreSqlite) CreateOne(ctx context.Context, arg *params.InvoiceCreate) (int64, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, store.ErrInsertFailed
	}

	defer tx.Rollback()

	insertInvoice := `
	INSERT INTO invoices(project_id, is_invoice, status, total, discount, tax, currency_code, note)
	VALUES(?, ?, (SELECT id FROM invoice_statuses WHERE name = ?), ?, ?, ?, ?, ?) RETURNING id
	`

	var invoiceId int64
	err = tx.QueryRowContext(ctx, insertInvoice,
		arg.ProjectId,
		arg.IsInvoice,
		arg.Status,
		arg.Total.String(),
		arg.Discount.String(),
		arg.Tax.String(),
		arg.CurrencyCode,
		arg.Note).Scan(&invoiceId)

	if err != nil {

		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return 0, store.ErrForeignKeyViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
				return 0, store.ErrNotNullViolation
			}
		}

		return 0, store.ErrInsertFailed
	}

	insertItem := `
	INSERT INTO invoice_items(
		invoice_id, description, qty, unit_price, unit_tax, tax_type, unit_discount, discount_type, tax, discount, total)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	for _, item := range arg.Items {
		if _, err = tx.ExecContext(ctx, insertItem,
			invoiceId,
			item.Description,
			item.Qty.String(),
			item.UnitPrice.String(),
			item.UnitTax.String(),
			item.TaxType,
			item.UnitDiscount.String(),
			item.DiscountType,
			item.Tax.String(),
			item.Discount.String(),
			item.Total.String()); err != nil {

			if sqlite3Err, ok := err.(sqlite3.Error); ok {
				if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
					return 0, store.ErrNotNullViolation
				}

				if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintCheck {
					return 0, store.ErrCheckViolation
				}
			}
			return 0, store.ErrInsertFailed
		}
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
func (q *InvoiceStoreSqlite) GetInvoiceStatsByProject(
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
	WHERE i.project_id = ? 
	`)

	countQuery.WriteString(`
	SELECT COUNT(i.id) FROM invoices i
	JOIN invoice_statuses ins ON ins.id = i.status
	WHERE i.project_id = ? 
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

	if len(arg.Status) > 0 {
		query.WriteString(" AND ins.name = ?")
		countQuery.WriteString(" AND ins.name = ?")

		queryArgs = append(queryArgs, arg.Status)
		countQueryArgs = append(countQueryArgs, arg.Status)
	}

	query.WriteString(" LIMIT ? OFFSET ?")
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
func (q *InvoiceStoreSqlite) GetWithProjectByUserId(
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
	WHERE pu.user_id = ?
	`)

	countQuery.WriteString(`
	SELECT COUNT(i.id)
	FROM project_users pu
	JOIN projects p ON p.id = pu.project_id
	JOIN invoices i ON i.project_id = pu.project_id
	JOIN invoice_statuses ins ON ins.id = i.status
	WHERE pu.user_id = ?
	`)

	queryArgs := []any{userId}
	countQueryArgs := []any{userId}

	if arg.Type == params.InvoiceTypeInvoice {
		query.WriteString(" AND i.is_invoice = true")
		countQuery.WriteString(" AND i.is_invoice = true")
	}

	if arg.Type == params.InvoiceTypeQuote {
		query.WriteString(" AND i.is_invoice = false")
		countQuery.WriteString(" AND i.is_invoice = false")
	}

	if len(arg.Keyword) > 0 {
		query.WriteString(" AND p.name LIKE ?")
		countQuery.WriteString(" AND p.name LIKE ?")

		keyword := "%" + arg.Keyword + "%"

		queryArgs = append(queryArgs, keyword)
		countQueryArgs = append(countQueryArgs, keyword)
	}

	if len(arg.Status) > 0 {
		query.WriteString(" AND ins.name = ?")
		countQuery.WriteString(" AND ins.name = ?")

		queryArgs = append(queryArgs, arg.Status)
		countQueryArgs = append(countQueryArgs, arg.Status)
	}

	query.WriteString(" LIMIT ? OFFSET ?")
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
func (q *InvoiceStoreSqlite) GetWithDetailsById(
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
	WHERE i.id = ?
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
	WHERE invoice_id = ?
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
