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
	//   - userId: ID of the user creating the invoice
	//   - arg: pointer to InvoiceCreateParams
	//
	// Returns:
	//   - int64: id of new invoice
	//   - store.ErrInsertFailed: if create fails
	// 	 - store.ErrForeignKeyViolation: if foreign key is invalid
	// 	 - store.ErrNotNullViolation: if required field is missing
	//	 - store.ErrCheckViolation: if check constraint fails
	CreateOne(ctx context.Context, userId int64, arg *params.InvoiceCreate) (int64, error)

	// GetSummaryByProjectId returns the total number of invoices/quotes
	// found and a list of invoices for
	// a specified project using the project ID.
	// The search criteria parameter can be used to filter results
	// by the type (invoice/quote), limit and offset results.
	//
	// If any error occurs, store.ErrQueryFailed is returned.
	GetSummaryByProjectId(ctx context.Context, projectId int64, arg *params.InvoiceSearch) ([]agg.Invoice, error)

	CountSummaryByProjectId(ctx context.Context, projectId int64, arg *params.InvoiceSearch) (int64, error)

	// GetWithProjectByUserId returns the total number of invoices/quotes
	// found and a list of invoices with project name for all projects
	// that belong to the specified user.
	// The search criteria parameter can be used to filter results
	// by a keyword, the type (invoice/quote), limit and offset results.
	//
	// If any error occurs, store.ErrQueryFailed is returned.
	GetWithProjectByUserId(
		ctx context.Context,
		userId int64,
		arg *params.InvoiceSearch) (*agg.WithCount[agg.InvoiceWithProject], error)

	GetSummaryById(ctx context.Context, invoiceId int64) (*agg.Invoice, error)

	GetItemsByInvoiceId(ctx context.Context, invoiceId int64) ([]agg.InvoiceItem, error)

	GetHistoryByInvoiceId(ctx context.Context, invoiceId int64) ([]agg.InvoiceHistory, error)

	// AcceptById marks a specified invoice as accepted.
	// The invoice is specified by invoice ID.
	// Returns a boolean and an error.
	//
	// If the returned boolean is:
	// 	- true: the invoice is already accepted
	//  - false: successfully accepted the invoice
	//
	// If any error occurs, store.ErrUpdateFailed is returned.
	AcceptById(ctx context.Context, userId int64, invoiceId int64) (bool, error)

	// RejectById marks a specified invoice as rejected.
	// The invoice is specified by invoice ID.
	// Returns a boolean and an error.
	//
	// If the returned boolean is:
	// 	- true: the invoice is already rejected
	//  - false: successfully rejected the invoice
	//
	// If any error occurs, store.ErrUpdateFailed is returned.
	RejectById(ctx context.Context, userId int64, invoiceId int64) (bool, error)

	// CancelById marks a specified invoice or a quote as cancelled.
	// The invoice/quote is specified by invoice ID.
	// Returns a boolean and an error.
	//
	// If the returned boolean is:
	// 	- true: the invoice or quote is already cancelled
	//  - false: successfully cancelled the invoice or quote
	//
	// If any error occurs, store.ErrUpdateFailed is returned.
	CancelById(ctx context.Context, userId int64, invoiceId int64) (bool, error)

	// PayById marks a specified invoice as paid.
	// The invoice is specified by invoice ID.
	// Returns a boolean and an error.
	//
	// If the returned boolean is:
	// 	- true: the invoice is already paid
	//  - false: successfully marked the invoice as paid
	//
	// Errors:
	// 	- if the provided ID points to a quote, store.ErrUnexpectedType is returned.
	// 	- If any other error occurs, store.ErrUpdateFailed is returned.
	PayById(ctx context.Context, userId int64, invoiceId int64) (bool, error)

	// QuoteToInvoice converts a quote to an invoice.
	// The quote is specified by quote ID.
	// Returns a boolean and an error.
	//
	// If the returned boolean is:
	// 	- true: the quote is already converted
	//  - false: successfully converted to an invoice
	//
	// Errors:
	// 	- if the quote status is invalid, store.ErrUnexpectedType is returned.
	// 	- If any other error occurs, store.ErrUpdateFailed is returned.
	QuoteToInvoice(ctx context.Context, userId int64, quoteId int64) (bool, error)
}
