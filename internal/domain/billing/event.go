package billing

import (
	"mizu/internal/domain/common"
	"mizu/internal/domain/project"
	"time"
)

var (
	EventTypeInvoiceCreated       common.EventType = "invoice.created"
	EventTypeInvoiceStatusChanged common.EventType = "invoice.status.changed"
	EventTypeInvoiceConverted     common.EventType = "invoice.converted"
)

type InvoiceCreatedEvent struct {
	InvoiceID    InvoiceID
	ProjectID    project.ProjectID
	SubTotal     Decimal
	CurrencyCode CurrencyCode
	IsInvoice    bool
	OccurredAt   time.Time
}

func (e InvoiceCreatedEvent) EventType() common.EventType {
	return EventTypeInvoiceCreated
}

type InvoiceStatusChangedEvent struct {
	InvoiceID    InvoiceID
	ProjectID    project.ProjectID
	SubTotal     Decimal
	CurrencyCode CurrencyCode
	Status       Status
	IsInvoice    bool
	OccurredAt   time.Time
}

func (e InvoiceStatusChangedEvent) EventType() common.EventType {
	return EventTypeInvoiceStatusChanged
}

type InvoiceConvertedEvent struct {
	InvoiceID    InvoiceID
	ProjectID    project.ProjectID
	SubTotal     Decimal
	CurrencyCode CurrencyCode
	IsInvoice    bool
	OccurredAt   time.Time
}

func (e InvoiceConvertedEvent) EventType() common.EventType {
	return EventTypeInvoiceConverted
}
