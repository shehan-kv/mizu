package aggregates

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

type InvoiceDetails struct {
	Id           int64
	ProjectId    int64
	ProjectName  string
	IsInvoice    bool
	Status       string
	IssuedAt     time.Time
	DueAt        *time.Time
	Total        *apd.Decimal
	Discount     *apd.Decimal
	Tax          *apd.Decimal
	CurrencyCode string
	Note         *string
	Items        []InvoiceItem
}
