package store

import (
	"context"
	agg "mizu/internal/db/models/aggregates"
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
	CreateOne(ctx context.Context, arg *params.InvoiceCreate) (int64, error)

	// GetInvoiceStatsByProject returns the total number of invoices/quotes
	// found and a list of invoices with status for
	// a specified project using the project ID.
	// The search criteria parameter can be used to filter results
	// by the type (invoice/quote), limit and offset results.
	//
	// If any error occurs, store.ErrQueryFailed is returned.
	GetInvoiceStatsByProject(
		ctx context.Context,
		projectId int64,
		arg *params.InvoiceSearch) (*agg.WithCount[agg.InvoiceWithStatus], error)
}
