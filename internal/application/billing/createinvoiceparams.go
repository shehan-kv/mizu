package billing

import "time"

type CreateInvoiceParams struct {
	ActorID      string
	ProjectID    string
	CurrencyCode string
	Note         *string
	DueDate      *time.Time
	IsInvoice    bool
	Items        []CreateItemParams
}
