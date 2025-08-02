package postgres

import (
	"context"
	"database/sql"
	"mizu/internal/db/params"
	"mizu/internal/db/store"

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
func (q *InvoiceStorePostgres) CreateOne(ctx context.Context, arg *params.InvoiceCreateParams) (int64, error) {

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
