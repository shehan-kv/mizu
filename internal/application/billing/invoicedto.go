package billing

import (
	"time"
)

type InvoiceDTO struct {
	ID           string
	ProjectID    string
	ProjectName  string
	IsInvoice    bool
	Status       string
	DueAt        *time.Time
	CurrencyCode string
	Note         *string
	Items        []InvoiceItemDTO

	CreatedAt time.Time
	UpdatedAt time.Time

	TotalTax      string
	TotalDiscount string
	SubTotal      string
}
