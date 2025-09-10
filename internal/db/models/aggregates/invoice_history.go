package aggregates

import "time"

type InvoiceHistory struct {
	Id         int64
	UserId     int64
	FirstName  string
	LastName   string
	Image      *string
	Title      *string
	Role       string
	Event      string
	RecordedAt time.Time
	IsInvoice  bool
	LastStatus *string
	NewStatus  *string
}
