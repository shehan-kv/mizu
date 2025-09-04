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
func (q *InvoiceStorePostgres) CreateOne(ctx context.Context, arg *params.InvoiceCreate) (int64, error) {

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
