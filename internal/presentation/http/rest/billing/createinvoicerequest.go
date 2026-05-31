package billing

import "time"

type CreateInvoiceRequest struct {
	CurrencyCode string              `json:"currencyCode"`
	Note         *string             `json:"note"`
	DueDate      *time.Time          `json:"dueDate"`
	IsInvoice    bool                `json:"isInvoice"`
	Items        []CreateItemRequest `json:"items"`
}
