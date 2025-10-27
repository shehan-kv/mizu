package aggregates

import "github.com/cockroachdb/apd/v3"

type InvoiceSum struct {
	CurrencyCode string
	Count        int64
	Sum          *apd.Decimal
}
