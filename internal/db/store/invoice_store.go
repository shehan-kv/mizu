package store

import (
	"context"
	"mizu/internal/db/params"
)

type InvoiceStore interface {

	// Creates an invoice with invoice items
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - arg: pointer to InvoiceCreateParams
	//
	// Returns:
	//   - int64: id of new invoice
	//   - store.ErrInsertFailed: if create fails
	// 	 - store.ErrForeignKeyViolation: if foreign key is invalid
	// 	 - store.ErrNotNullViolation: if required field is missing
	//	 - store.ErrCheckViolation: if check constraint fails
	CreateOne(ctx context.Context, arg *params.InvoiceCreateParams) (int64, error)
}
