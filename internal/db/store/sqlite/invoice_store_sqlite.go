package sqlite

import (
	"context"
	"database/sql"
	"mizu/internal/db/params"
	"mizu/internal/db/store"

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
