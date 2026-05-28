package billing

import (
	"mizu/internal/domain/common"
	"mizu/internal/domain/project"
	"time"
)

type Invoice struct {
	id        InvoiceID
	projectID project.ProjectID
	isInvoice bool
	status    Status
	dueAt     *time.Time
	currency  Currency
	note      *string
	items     []Item

	events []common.Event

	//precomputed values
	totalTax      Decimal
	totalDiscount Decimal
	subTotal      Decimal

	version   int
	createdAt time.Time
	updatedAt time.Time
}

func NewInvoice(
	id InvoiceID,
	projectID project.ProjectID,
	isInvoice bool,
	dueAt *time.Time,
	currency Currency,
	note *string,
	items []Item,
	now time.Time,
) (*Invoice, error) {

	if len(items) == 0 {
		return nil, ErrBillingInvoiceMustHaveItems
	}

	totalTax, err := NewDecimal("0")
	if err != nil {
		return nil, err
	}
	totalDiscount, err := NewDecimal("0")
	if err != nil {
		return nil, err
	}
	subTotal, err := NewDecimal("0")
	if err != nil {
		return nil, err
	}

	for i := range items {
		newTax, err := totalTax.Add(items[i].LineTax())
		if err != nil {
			return nil, err
		}

		newDiscount, err := totalDiscount.Add(items[i].LineDiscount())
		if err != nil {
			return nil, err
		}

		newSubTotal, err := subTotal.Add(items[i].LineTotal())
		if err != nil {
			return nil, err
		}

		totalTax = newTax
		totalDiscount = newDiscount
		subTotal = newSubTotal
	}

	inv := Invoice{
		id:        id,
		projectID: projectID,
		isInvoice: isInvoice,
		status:    StatusPending,
		dueAt:     dueAt,
		currency:  currency,
		note:      note,
		items:     items,

		totalTax:      totalTax,
		totalDiscount: totalDiscount,
		subTotal:      subTotal,

		version:   1,
		createdAt: now,
		updatedAt: now,
	}

	inv.events = append(inv.events, InvoiceCreatedEvent{
		InvoiceID:    id,
		ProjectID:    projectID,
		IsInvoice:    isInvoice,
		SubTotal:     subTotal,
		CurrencyCode: currency.code,
		OccurredAt:   now,
	})

	return &inv, nil
}

func RestoreInvoice(
	id InvoiceID,
	projectID project.ProjectID,
	isInvoice bool,
	dueAt *time.Time,
	currency Currency,
	note *string,
	items []Item,
	totalTax Decimal,
	totalDiscount Decimal,
	subTotal Decimal,
	version int,
	createdAt time.Time,
	updatedAt time.Time,
) Invoice {

	return Invoice{
		id:        id,
		projectID: projectID,
		isInvoice: isInvoice,
		status:    StatusPending,
		dueAt:     dueAt,
		currency:  currency,
		note:      note,
		items:     items,

		totalTax:      totalTax,
		totalDiscount: totalDiscount,
		subTotal:      subTotal,

		version:   version,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (inv *Invoice) ID() InvoiceID {
	return inv.id
}

func (inv *Invoice) ProjectID() project.ProjectID {
	return inv.projectID
}

func (inv *Invoice) IsInvoice() bool {
	return inv.isInvoice
}

func (inv *Invoice) IsQuote() bool {
	return !inv.isInvoice
}

func (inv *Invoice) Status() Status {
	return inv.status
}

func (inv *Invoice) DueAt() *time.Time {
	return inv.dueAt
}

func (inv *Invoice) Currency() Currency {
	return inv.currency
}

func (inv *Invoice) Note() *string {
	return inv.note
}

func (inv *Invoice) Items() []Item {
	items := make([]Item, len(inv.items))
	copy(items, inv.items)
	return items
}

func (inv *Invoice) Reject(now time.Time) error {
	if inv.status != StatusPending {
		return ErrBillingRejectRequiresPending
	}
	inv.status = StatusRejected
	inv.updatedAt = now

	inv.events = append(inv.events, InvoiceStatusChangedEvent{
		InvoiceID:    inv.id,
		ProjectID:    inv.projectID,
		Status:       inv.status,
		CurrencyCode: inv.currency.code,
		SubTotal:     inv.subTotal,
		IsInvoice:    inv.isInvoice,
		OccurredAt:   now,
	})

	return nil
}

func (inv *Invoice) Accept(now time.Time) error {
	if inv.status != StatusPending {
		return ErrBillingAcceptRequiresPending
	}
	inv.status = StatusAccepted
	inv.updatedAt = now

	inv.events = append(inv.events, InvoiceStatusChangedEvent{
		InvoiceID:    inv.id,
		ProjectID:    inv.projectID,
		Status:       inv.status,
		CurrencyCode: inv.currency.code,
		SubTotal:     inv.subTotal,
		IsInvoice:    inv.isInvoice,
		OccurredAt:   now,
	})

	return nil
}

func (inv *Invoice) Cancel(now time.Time) error {
	if inv.status != StatusPending && inv.status != StatusAccepted {
		return ErrBillingCancelInvalidStatus
	}
	inv.status = StatusCancelled
	inv.updatedAt = now

	inv.events = append(inv.events, InvoiceStatusChangedEvent{
		InvoiceID:    inv.id,
		ProjectID:    inv.projectID,
		Status:       inv.status,
		CurrencyCode: inv.currency.code,
		SubTotal:     inv.subTotal,
		IsInvoice:    inv.isInvoice,
		OccurredAt:   now,
	})

	return nil
}

func (inv *Invoice) Pay(now time.Time) error {
	if inv.status != StatusAccepted {
		return ErrBillingPayRequiresAccepted
	}
	inv.status = StatusPaid
	inv.updatedAt = now

	inv.events = append(inv.events, InvoiceStatusChangedEvent{
		InvoiceID:    inv.id,
		ProjectID:    inv.projectID,
		Status:       inv.status,
		CurrencyCode: inv.currency.code,
		SubTotal:     inv.subTotal,
		IsInvoice:    inv.isInvoice,
		OccurredAt:   now,
	})

	return nil
}

func (inv *Invoice) ConvertToInvoice(now time.Time) error {
	if inv.isInvoice {
		return ErrBillingAlreadyAnInvoice
	}
	if inv.status != StatusPending {
		return ErrBillingConvertRequiresPending
	}

	inv.isInvoice = true
	inv.updatedAt = now

	inv.events = append(inv.events, InvoiceConvertedEvent{
		InvoiceID:    inv.id,
		ProjectID:    inv.projectID,
		SubTotal:     inv.subTotal,
		CurrencyCode: inv.currency.code,
		IsInvoice:    inv.isInvoice,
		OccurredAt:   now,
	})

	return nil
}

func (inv *Invoice) Version() int {
	return inv.version
}

func (inv *Invoice) CreatedAt() time.Time {
	return inv.createdAt
}

func (inv *Invoice) UpdatedAt() time.Time {
	return inv.updatedAt
}

func (inv *Invoice) TotalTax() Decimal {
	return inv.totalTax
}

func (inv *Invoice) TotalDiscount() Decimal {
	return inv.totalDiscount
}

func (inv *Invoice) SubTotal() Decimal {
	return inv.subTotal
}

func (inv *Invoice) PullEvents() []common.Event {
	events := inv.events
	inv.events = nil

	return events
}

func (inv *Invoice) Equals(other *Invoice) bool {
	return inv.id == other.id
}
