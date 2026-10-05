package billing

import "time"

type InvoiceOverviewDTO struct {
	ID            string
	ProjectID     string
	ProjectName   string
	IsInvoice     bool
	Status        string
	DueAt         *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CurrencyCode  string
	Note          *string
	TotalTax      string
	TotalDiscount string
	SubTotal      string
}
