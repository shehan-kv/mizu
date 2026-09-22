package billing

import (
	"testing"
	"time"

	"mizu/internal/domain/project"
)

func newTestInvoiceItem(t *testing.T, description string, unitPrice string) Item {
	t.Helper()

	return newTestItem(
		t,
		description,
		"2",
		unitPrice,
		"10",
		DiscountTypeFixed,
		"5",
		TaxTypeFixed,
	)
}

func newTestInvoice(t *testing.T, now time.Time, isInvoice bool) *Invoice {
	t.Helper()

	invoiceID, err := NewInvoiceID("invoice-1")
	if err != nil {
		t.Fatal(err)
	}

	projectID, err := project.NewProjectID("project-1")
	if err != nil {
		t.Fatal(err)
	}

	name, err := NewCurrencyName("US Dollar")
	if err != nil {
		t.Fatal(err)
	}

	symbol, err := NewCurrencySymbol("$")
	if err != nil {
		t.Fatal(err)
	}

	code, err := NewCurrencyCode("USD")
	if err != nil {
		t.Fatal(err)
	}

	decimals, err := NewCurrencyDecimals(2)
	if err != nil {
		t.Fatal(err)
	}

	currency := NewCurrency(name, symbol, code, decimals)

	item := newTestInvoiceItem(t, "Development", "100")

	invoice, err := NewInvoice(
		invoiceID,
		projectID,
		isInvoice,
		nil,
		currency,
		nil,
		[]Item{item},
		now,
	)
	if err != nil {
		t.Fatal(err)
	}

	return invoice
}

func TestNewInvoice(t *testing.T) {
	now := time.Now()

	invoice := newTestInvoice(t, now, true)

	if invoice.ID().String() != "invoice-1" {
		t.Fatalf("expected invoice-1, got %s", invoice.ID())
	}

	if invoice.ProjectID().String() != "project-1" {
		t.Fatalf("expected project-1, got %s", invoice.ProjectID())
	}

	if !invoice.IsInvoice() {
		t.Fatal("expected invoice")
	}

	if invoice.IsQuote() {
		t.Fatal("expected invoice not to be quote")
	}

	if invoice.Status() != StatusPending {
		t.Fatalf("expected pending status, got %s", invoice.Status())
	}

	if invoice.Version() != 1 {
		t.Fatalf("expected version 1, got %d", invoice.Version())
	}

	if !invoice.CreatedAt().Equal(now) {
		t.Fatalf("expected createdAt %v, got %v", now, invoice.CreatedAt())
	}

	if !invoice.UpdatedAt().Equal(now) {
		t.Fatalf("expected updatedAt %v, got %v", now, invoice.UpdatedAt())
	}

	if len(invoice.Items()) != 1 {
		t.Fatalf("expected 1 item, got %d", len(invoice.Items()))
	}

	expectedTax := mustDecimal(t, "10")
	if !invoice.TotalTax().Equals(expectedTax) {
		t.Fatalf("expected total tax 10, got %s", invoice.TotalTax())
	}

	expectedDiscount := mustDecimal(t, "20")
	if !invoice.TotalDiscount().Equals(expectedDiscount) {
		t.Fatalf("expected total discount 20, got %s", invoice.TotalDiscount())
	}

	expectedSubtotal := mustDecimal(t, "190")
	if !invoice.SubTotal().Equals(expectedSubtotal) {
		t.Fatalf("expected subtotal 190, got %s", invoice.SubTotal())
	}

	events := invoice.PullEvents()

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event, ok := events[0].(InvoiceCreatedEvent)
	if !ok {
		t.Fatalf("expected InvoiceCreatedEvent, got %T", events[0])
	}

	if event.InvoiceID != invoice.ID() {
		t.Fatalf("expected event invoice ID %q, got %q", invoice.ID(), event.InvoiceID)
	}

	if event.ProjectID != invoice.ProjectID() {
		t.Fatalf("expected event project ID %q, got %q", invoice.ProjectID(), event.ProjectID)
	}

	if !event.SubTotal.Equals(invoice.SubTotal()) {
		t.Fatalf("expected event subtotal %s, got %s", invoice.SubTotal(), event.SubTotal)
	}

	if event.CurrencyCode != invoice.Currency().Code() {
		t.Fatalf("expected currency code %q, got %q", invoice.Currency().Code(), event.CurrencyCode)
	}

	if !event.IsInvoice {
		t.Fatal("expected event IsInvoice to be true")
	}

	if !event.OccurredAt.Equal(now) {
		t.Fatalf("expected event occurredAt %v, got %v", now, event.OccurredAt)
	}

	if event.EventType() != EventTypeInvoiceCreated {
		t.Fatalf("expected event type %q, got %q", EventTypeInvoiceCreated, event.EventType())
	}
}

func TestNewInvoiceRequiresItems(t *testing.T) {
	now := time.Now()

	invoiceID, err := NewInvoiceID("invoice-1")
	if err != nil {
		t.Fatal(err)
	}

	projectID, err := project.NewProjectID("project-1")
	if err != nil {
		t.Fatal(err)
	}

	currencyName, err := NewCurrencyName("US Dollar")
	if err != nil {
		t.Fatal(err)
	}

	currencySymbol, err := NewCurrencySymbol("$")
	if err != nil {
		t.Fatal(err)
	}

	currencyCode, err := NewCurrencyCode("USD")
	if err != nil {
		t.Fatal(err)
	}

	currencyDecimals, err := NewCurrencyDecimals(2)
	if err != nil {
		t.Fatal(err)
	}

	currency := NewCurrency(
		currencyName,
		currencySymbol,
		currencyCode,
		currencyDecimals,
	)

	_, err = NewInvoice(
		invoiceID,
		projectID,
		true,
		nil,
		currency,
		nil,
		nil,
		now,
	)

	if err != ErrBillingInvoiceMustHaveItems {
		t.Fatalf("expected %v, got %v", ErrBillingInvoiceMustHaveItems, err)
	}
}

func TestNewInvoiceCalculatesTotalsAcrossItems(t *testing.T) {
	now := time.Now()

	invoiceID, err := NewInvoiceID("invoice-1")
	if err != nil {
		t.Fatal(err)
	}

	projectID, err := project.NewProjectID("project-1")
	if err != nil {
		t.Fatal(err)
	}

	currencyName, err := NewCurrencyName("US Dollar")
	if err != nil {
		t.Fatal(err)
	}

	currencySymbol, err := NewCurrencySymbol("$")
	if err != nil {
		t.Fatal(err)
	}

	currencyCode, err := NewCurrencyCode("USD")
	if err != nil {
		t.Fatal(err)
	}

	currencyDecimals, err := NewCurrencyDecimals(2)
	if err != nil {
		t.Fatal(err)
	}

	currency := NewCurrency(
		currencyName,
		currencySymbol,
		currencyCode,
		currencyDecimals,
	)

	item1 := newTestInvoiceItem(t, "Development", "100")
	item2 := newTestInvoiceItem(t, "Testing", "50")

	invoice, err := NewInvoice(
		invoiceID,
		projectID,
		true,
		nil,
		currency,
		nil,
		[]Item{item1, item2},
		now,
	)
	if err != nil {
		t.Fatal(err)
	}

	expectedTax := mustDecimal(t, "20")
	if !invoice.TotalTax().Equals(expectedTax) {
		t.Fatalf("expected total tax %s, got %s", expectedTax, invoice.TotalTax())
	}

	expectedDiscount := mustDecimal(t, "40")
	if !invoice.TotalDiscount().Equals(expectedDiscount) {
		t.Fatalf("expected total discount %s, got %s", expectedDiscount, invoice.TotalDiscount())
	}

	expectedSubtotal := mustDecimal(t, "280")
	if !invoice.SubTotal().Equals(expectedSubtotal) {
		t.Fatalf("expected subtotal %s, got %s", expectedSubtotal, invoice.SubTotal())
	}
}

func TestRestoreInvoice(t *testing.T) {
	createdAt := time.Now()
	updatedAt := createdAt.Add(time.Second)

	invoiceID, err := NewInvoiceID("invoice-1")
	if err != nil {
		t.Fatal(err)
	}

	projectID, err := project.NewProjectID("project-1")
	if err != nil {
		t.Fatal(err)
	}

	currencyName, err := NewCurrencyName("US Dollar")
	if err != nil {
		t.Fatal(err)
	}

	currencySymbol, err := NewCurrencySymbol("$")
	if err != nil {
		t.Fatal(err)
	}

	currencyCode, err := NewCurrencyCode("USD")
	if err != nil {
		t.Fatal(err)
	}

	currencyDecimals, err := NewCurrencyDecimals(2)
	if err != nil {
		t.Fatal(err)
	}

	currency := NewCurrency(
		currencyName,
		currencySymbol,
		currencyCode,
		currencyDecimals,
	)

	item := newTestInvoiceItem(t, "Development", "100")
	totalTax := mustDecimal(t, "10")
	totalDiscount := mustDecimal(t, "20")
	subTotal := mustDecimal(t, "190")

	invoice := RestoreInvoice(
		invoiceID,
		projectID,
		true,
		StatusAccepted,
		nil,
		currency,
		nil,
		[]Item{item},
		totalTax,
		totalDiscount,
		subTotal,
		4,
		createdAt,
		updatedAt,
	)

	if invoice.Status() != StatusAccepted {
		t.Fatalf("expected accepted status, got %s", invoice.Status())
	}

	if invoice.Version() != 4 {
		t.Fatalf("expected version 4, got %d", invoice.Version())
	}

	if !invoice.CreatedAt().Equal(createdAt) {
		t.Fatalf("expected createdAt %v, got %v", createdAt, invoice.CreatedAt())
	}

	if !invoice.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updatedAt %v, got %v", updatedAt, invoice.UpdatedAt())
	}

	if !invoice.TotalTax().Equals(totalTax) {
		t.Fatal("unexpected total tax")
	}

	if !invoice.TotalDiscount().Equals(totalDiscount) {
		t.Fatal("unexpected total discount")
	}

	if !invoice.SubTotal().Equals(subTotal) {
		t.Fatal("unexpected subtotal")
	}

	if events := invoice.PullEvents(); len(events) != 0 {
		t.Fatalf("restored invoice should have no events, got %d", len(events))
	}
}

func TestInvoiceReject(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, false)

	updatedAt := now.Add(time.Second)

	if err := invoice.Reject(updatedAt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.Status() != StatusRejected {
		t.Fatalf("expected rejected status, got %s", invoice.Status())
	}

	if !invoice.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updatedAt %v, got %v", updatedAt, invoice.UpdatedAt())
	}

	events := invoice.PullEvents()

	if len(events) != 2 {
		t.Fatalf("expected creation and status events, got %d", len(events))
	}

	event, ok := events[1].(InvoiceStatusChangedEvent)
	if !ok {
		t.Fatalf("expected InvoiceStatusChangedEvent, got %T", events[1])
	}

	if event.Status != StatusRejected {
		t.Fatalf("expected rejected status in event, got %s", event.Status)
	}

	if !event.OccurredAt.Equal(updatedAt) {
		t.Fatalf("expected event occurredAt %v, got %v", updatedAt, event.OccurredAt)
	}

	if event.EventType() != EventTypeInvoiceStatusChanged {
		t.Fatalf("expected event type %q, got %q", EventTypeInvoiceStatusChanged, event.EventType())
	}
}

func TestInvoiceRejectRequiresPending(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, true)

	if err := invoice.Accept(now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	if err := invoice.Reject(now.Add(2 * time.Second)); err != ErrBillingRejectRequiresPending {
		t.Fatalf("expected %v, got %v", ErrBillingRejectRequiresPending, err)
	}
}

func TestInvoiceAccept(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, true)

	updatedAt := now.Add(time.Second)

	if err := invoice.Accept(updatedAt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.Status() != StatusAccepted {
		t.Fatalf("expected accepted status, got %s", invoice.Status())
	}

	if !invoice.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updatedAt %v, got %v", updatedAt, invoice.UpdatedAt())
	}

	events := invoice.PullEvents()

	if len(events) != 2 {
		t.Fatalf("expected creation and status events, got %d", len(events))
	}

	event, ok := events[1].(InvoiceStatusChangedEvent)
	if !ok {
		t.Fatalf("expected InvoiceStatusChangedEvent, got %T", events[1])
	}

	if event.Status != StatusAccepted {
		t.Fatalf("expected accepted status in event, got %s", event.Status)
	}
}

func TestInvoiceAcceptRequiresPending(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, true)

	if err := invoice.Accept(now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	if err := invoice.Accept(now.Add(2 * time.Second)); err != ErrBillingAcceptRequiresPending {
		t.Fatalf("expected %v, got %v", ErrBillingAcceptRequiresPending, err)
	}
}

func TestInvoiceCancelFromPending(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, true)

	updatedAt := now.Add(time.Second)

	if err := invoice.Cancel(updatedAt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.Status() != StatusCancelled {
		t.Fatalf("expected cancelled status, got %s", invoice.Status())
	}

	if !invoice.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updatedAt %v, got %v", updatedAt, invoice.UpdatedAt())
	}
}

func TestInvoiceCancelFromAccepted(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, true)

	if err := invoice.Accept(now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	updatedAt := now.Add(2 * time.Second)

	if err := invoice.Cancel(updatedAt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.Status() != StatusCancelled {
		t.Fatalf("expected cancelled status, got %s", invoice.Status())
	}
}

func TestInvoiceCancelRejectsInvalidStatus(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, true)

	if err := invoice.Reject(now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	if err := invoice.Cancel(now.Add(2 * time.Second)); err != ErrBillingCancelInvalidStatus {
		t.Fatalf("expected %v, got %v", ErrBillingCancelInvalidStatus, err)
	}
}

func TestInvoicePay(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, true)

	if err := invoice.Accept(now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	updatedAt := now.Add(2 * time.Second)

	if err := invoice.Pay(updatedAt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.Status() != StatusPaid {
		t.Fatalf("expected paid status, got %s", invoice.Status())
	}

	if !invoice.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updatedAt %v, got %v", updatedAt, invoice.UpdatedAt())
	}

	events := invoice.PullEvents()

	if len(events) != 3 {
		t.Fatalf("expected creation, accepted, and paid events, got %d", len(events))
	}

	event, ok := events[2].(InvoiceStatusChangedEvent)
	if !ok {
		t.Fatalf("expected InvoiceStatusChangedEvent, got %T", events[2])
	}

	if event.Status != StatusPaid {
		t.Fatalf("expected paid status in event, got %s", event.Status)
	}
}

func TestInvoicePayRequiresAccepted(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, true)

	if err := invoice.Pay(now.Add(time.Second)); err != ErrBillingPayRequiresAccepted {
		t.Fatalf("expected %v, got %v", ErrBillingPayRequiresAccepted, err)
	}
}

func TestInvoiceConvertToInvoice(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, false)

	updatedAt := now.Add(time.Second)

	if err := invoice.ConvertToInvoice(updatedAt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !invoice.IsInvoice() {
		t.Fatal("expected invoice after conversion")
	}

	if invoice.IsQuote() {
		t.Fatal("expected converted quote not to be a quote")
	}

	if !invoice.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updatedAt %v, got %v", updatedAt, invoice.UpdatedAt())
	}

	events := invoice.PullEvents()

	if len(events) != 2 {
		t.Fatalf("expected creation and conversion events, got %d", len(events))
	}

	event, ok := events[1].(InvoiceConvertedEvent)
	if !ok {
		t.Fatalf("expected InvoiceConvertedEvent, got %T", events[1])
	}

	if !event.IsInvoice {
		t.Fatal("expected conversion event IsInvoice to be true")
	}

	if !event.OccurredAt.Equal(updatedAt) {
		t.Fatalf("expected event occurredAt %v, got %v", updatedAt, event.OccurredAt)
	}

	if event.EventType() != EventTypeInvoiceConverted {
		t.Fatalf("expected event type %q, got %q", EventTypeInvoiceConverted, event.EventType())
	}
}

func TestInvoiceConvertToInvoiceRejectsExistingInvoice(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, true)

	if err := invoice.ConvertToInvoice(now.Add(time.Second)); err != ErrBillingAlreadyAnInvoice {
		t.Fatalf("expected %v, got %v", ErrBillingAlreadyAnInvoice, err)
	}
}

func TestInvoiceConvertToInvoiceRequiresPending(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, false)

	if err := invoice.Accept(now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	if err := invoice.ConvertToInvoice(now.Add(2 * time.Second)); err != ErrBillingConvertRequiresPending {
		t.Fatalf("expected %v, got %v", ErrBillingConvertRequiresPending, err)
	}
}

func TestInvoicePullEventsClearsEvents(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, true)

	first := invoice.PullEvents()

	if len(first) != 1 {
		t.Fatalf("expected 1 event, got %d", len(first))
	}

	second := invoice.PullEvents()

	if len(second) != 0 {
		t.Fatalf("expected no events after pulling, got %d", len(second))
	}
}

func TestInvoiceItemsReturnsCopy(t *testing.T) {
	now := time.Now()
	invoice := newTestInvoice(t, now, true)

	items := invoice.Items()

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	items[0] = newTestInvoiceItem(t, "Replacement", "999")

	originalItems := invoice.Items()

	if originalItems[0].Description() != "Development" {
		t.Fatal("modifying returned items should not modify invoice items")
	}
}

func TestInvoiceEquals(t *testing.T) {
	now := time.Now()

	invoice1 := newTestInvoice(t, now, true)

	invoice2 := newTestInvoice(t, now.Add(time.Second), true)

	if !invoice1.Equals(invoice2) {
		t.Fatal("invoices with the same ID should be equal")
	}

	otherID, err := NewInvoiceID("invoice-2")
	if err != nil {
		t.Fatal(err)
	}

	projectID := invoice1.ProjectID()

	other := RestoreInvoice(
		otherID,
		projectID,
		true,
		StatusPending,
		nil,
		invoice1.Currency(),
		nil,
		invoice1.Items(),
		invoice1.TotalTax(),
		invoice1.TotalDiscount(),
		invoice1.SubTotal(),
		1,
		invoice1.CreatedAt(),
		invoice1.UpdatedAt(),
	)

	if invoice1.Equals(&other) {
		t.Fatal("invoices with different IDs should not be equal")
	}
}
