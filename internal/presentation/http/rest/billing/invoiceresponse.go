package billing

import "time"

type InvoiceResponse struct {
	ID            string                `json:"id"`
	ProjectID     string                `json:"projectId"`
	ProjectName   string                `json:"projectName"`
	IsInvoice     bool                  `json:"isInvoice"`
	Status        string                `json:"status"`
	DueAt         *time.Time            `json:"dueAt"`
	CurrencyCode  string                `json:"currencyCode"`
	Note          *string               `json:"note"`
	TotalTax      string                `json:"totalTax"`
	TotalDiscount string                `json:"totalDiscount"`
	SubTotal      string                `json:"subTotal"`
	Items         []InvoiceItemResponse `json:"items"`
	CreatedAt     time.Time             `json:"createdAt"`
	UpdatedAt     time.Time             `json:"updatedAt"`
}
