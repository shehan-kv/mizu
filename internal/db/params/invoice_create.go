package params

import "github.com/cockroachdb/apd/v3"

// Parameters of an invoice item
type InvoiceItem struct {
	Description  string
	Qty          *apd.Decimal
	UnitPrice    *apd.Decimal
	UnitDiscount *apd.Decimal
	DiscountType string
	UnitTax      *apd.Decimal
	TaxType      string
	Tax          *apd.Decimal
	Discount     *apd.Decimal
	Total        *apd.Decimal
}

// Parameters to create an invoice
type InvoiceCreate struct {
	ProjectId    int64
	IsInvoice    bool
	Status       string
	Total        *apd.Decimal
	Discount     *apd.Decimal
	Tax          *apd.Decimal
	CurrencyCode string
	Note         string
	Items        []InvoiceItem
}
