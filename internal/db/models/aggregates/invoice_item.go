package aggregates

import "github.com/cockroachdb/apd/v3"

type InvoiceItem struct {
	Id           int64
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
