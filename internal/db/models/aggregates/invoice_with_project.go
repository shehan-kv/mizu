package aggregates

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

type InvoiceWithProject struct {
	Id           int64
	ProjectId    int64
	ProjectName  string
	IsInvoice    bool
	IssuedAt     time.Time
	DueAt        *time.Time
	Total        *apd.Decimal
	CurrencyCode string
	Status       string
}
