package aggregates

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

type InvoiceWithStatus struct {
	Id           int64
	IsInvoice    bool
	IssuedAt     time.Time
	DueAt        *time.Time
	Total        *apd.Decimal
	CurrencyCode string
	Status       string
}
