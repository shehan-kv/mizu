package mailer

import "time"

type InvoiceEmail struct {
	Subject        string
	RecipientEmail string

	InvoiceID string
	ProjectID string

	Status       string
	CurrencyName string
	CurrencyCode string

	DueAt *time.Time
	Note  *string

	Items []InvoiceItem

	SubTotal      string
	TotalTax      string
	TotalDiscount string
}
