package eventhandler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"mizu/internal/application/eventbus"
	"mizu/internal/application/message/integration"
	"mizu/internal/domain/billing"
	"mizu/internal/domain/message"
)

func TestAddInvoiceStatusChangedMessage_Handle(t *testing.T) {
	t.Run("ignores unsupported event", func(t *testing.T) {
		deps := newTestDeps()

		handler := NewAddInvoiceStatusChangedMessage(
			newTestSystemMessagePublisher(deps),
		)

		err := handler.Handle(
			context.Background(),
			newTestUserCreatedEvent(t),
		)
		if err != nil {
			t.Fatalf("Handle() error = %v, want nil", err)
		}

		if deps.channelRepo.listByProjectCalls != 0 {
			t.Fatalf(
				"ListByProject() calls = %d, want 0",
				deps.channelRepo.listByProjectCalls,
			)
		}

		if deps.messageRepo.addCalls != 0 {
			t.Fatalf(
				"MessageRepository.Add() calls = %d, want 0",
				deps.messageRepo.addCalls,
			)
		}

		if deps.externalBus.publishCalls != 0 {
			t.Fatalf(
				"ExternalBus.Publish() calls = %d, want 0",
				deps.externalBus.publishCalls,
			)
		}
	})

	t.Run("publishes changed quote system message", func(t *testing.T) {
		deps := newTestDeps()

		projectID := newTestProjectID(t, "project-1")

		channel := newTestProjectChannel(
			t,
			"channel-1",
			projectID,
			"Project Channel",
			"member-1",
			"member-2",
		)

		deps.channelRepo.channels = []*message.Channel{channel}
		deps.idGen.ids = []string{"message-1"}

		event := newTestInvoiceStatusChangedEvent(t)
		event.ProjectID = projectID
		event.IsInvoice = false

		subTotal, err := billing.NewDecimal("100.00")
		if err != nil {
			t.Fatal(err)
		}
		event.SubTotal = subTotal

		handler := NewAddInvoiceStatusChangedMessage(
			newTestSystemMessagePublisher(deps),
		)

		err = handler.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if deps.channelRepo.listByProjectCalls != 1 {
			t.Fatalf(
				"ListByProject() calls = %d, want 1",
				deps.channelRepo.listByProjectCalls,
			)
		}

		if deps.messageRepo.addCalls != 1 {
			t.Fatalf(
				"MessageRepository.Add() calls = %d, want 1",
				deps.messageRepo.addCalls,
			)
		}

		if deps.externalBus.publishCalls != 1 {
			t.Fatalf(
				"ExternalBus.Publish() calls = %d, want 1",
				deps.externalBus.publishCalls,
			)
		}

		if len(deps.externalBus.events) != 1 {
			t.Fatalf(
				"published events = %d, want 1",
				len(deps.externalBus.events),
			)
		}

		published, ok := deps.externalBus.events[0].(integration.MessageBroadcast)
		if !ok {
			t.Fatalf(
				"published event type = %T, want integration.MessageBroadcast",
				deps.externalBus.events[0],
			)
		}

		var payload struct {
			Type         string `json:"type"`
			Status       string `json:"status"`
			DocumentType string `json:"documentType"`
			InvoiceID    string `json:"invoiceId"`
			CurrencyCode string `json:"currencyCode"`
			SubTotal     string `json:"subTotal"`
		}

		if err := json.Unmarshal([]byte(published.Content), &payload); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if payload.Type != billing.EventTypeInvoiceStatusChanged.String() {
			t.Fatalf(
				"type = %q, want %q",
				payload.Type,
				billing.EventTypeInvoiceStatusChanged.String(),
			)
		}

		if payload.Status != event.Status.String() {
			t.Fatalf(
				"status = %q, want %q",
				payload.Status,
				event.Status.String(),
			)
		}

		if payload.DocumentType != "quote" {
			t.Fatalf(
				"documentType = %q, want %q",
				payload.DocumentType,
				"quote",
			)
		}

		if payload.InvoiceID != event.InvoiceID.String() {
			t.Fatalf(
				"invoiceId = %q, want %q",
				payload.InvoiceID,
				event.InvoiceID.String(),
			)
		}

		if payload.CurrencyCode != event.CurrencyCode.String() {
			t.Fatalf(
				"currencyCode = %q, want %q",
				payload.CurrencyCode,
				event.CurrencyCode.String(),
			)
		}

		if payload.SubTotal != event.SubTotal.String() {
			t.Fatalf(
				"subTotal = %q, want %q",
				payload.SubTotal,
				event.SubTotal.String(),
			)
		}
	})

	t.Run("publishes changed invoice system message", func(t *testing.T) {
		deps := newTestDeps()

		projectID := newTestProjectID(t, "project-1")

		channel := newTestProjectChannel(
			t,
			"channel-1",
			projectID,
			"Project Channel",
			"member-1",
			"member-2",
		)

		deps.channelRepo.channels = []*message.Channel{channel}
		deps.idGen.ids = []string{"message-1"}

		event := newTestInvoiceStatusChangedEvent(t)
		event.ProjectID = projectID
		event.IsInvoice = true

		subTotal, err := billing.NewDecimal("100.00")
		if err != nil {
			t.Fatal(err)
		}
		event.SubTotal = subTotal

		handler := NewAddInvoiceStatusChangedMessage(
			newTestSystemMessagePublisher(deps),
		)

		err = handler.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if deps.messageRepo.addCalls != 1 {
			t.Fatalf(
				"MessageRepository.Add() calls = %d, want 1",
				deps.messageRepo.addCalls,
			)
		}

		if deps.externalBus.publishCalls != 1 {
			t.Fatalf(
				"ExternalBus.Publish() calls = %d, want 1",
				deps.externalBus.publishCalls,
			)
		}

		if len(deps.externalBus.events) != 1 {
			t.Fatalf(
				"published events = %d, want 1",
				len(deps.externalBus.events),
			)
		}

		published, ok := deps.externalBus.events[0].(integration.MessageBroadcast)
		if !ok {
			t.Fatalf(
				"published event type = %T, want integration.MessageBroadcast",
				deps.externalBus.events[0],
			)
		}

		var payload struct {
			Type         string `json:"type"`
			Status       string `json:"status"`
			DocumentType string `json:"documentType"`
			InvoiceID    string `json:"invoiceId"`
			CurrencyCode string `json:"currencyCode"`
			SubTotal     string `json:"subTotal"`
		}

		if err := json.Unmarshal([]byte(published.Content), &payload); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if payload.Type != billing.EventTypeInvoiceStatusChanged.String() {
			t.Fatalf(
				"type = %q, want %q",
				payload.Type,
				billing.EventTypeInvoiceStatusChanged.String(),
			)
		}

		if payload.Status != event.Status.String() {
			t.Fatalf(
				"status = %q, want %q",
				payload.Status,
				event.Status.String(),
			)
		}

		if payload.DocumentType != "invoice" {
			t.Fatalf(
				"documentType = %q, want %q",
				payload.DocumentType,
				"invoice",
			)
		}

		if payload.InvoiceID != event.InvoiceID.String() {
			t.Fatalf(
				"invoiceId = %q, want %q",
				payload.InvoiceID,
				event.InvoiceID.String(),
			)
		}

		if payload.CurrencyCode != event.CurrencyCode.String() {
			t.Fatalf(
				"currencyCode = %q, want %q",
				payload.CurrencyCode,
				event.CurrencyCode.String(),
			)
		}

		if payload.SubTotal != event.SubTotal.String() {
			t.Fatalf(
				"subTotal = %q, want %q",
				payload.SubTotal,
				event.SubTotal.String(),
			)
		}
	})

	t.Run("returns publisher error", func(t *testing.T) {
		deps := newTestDeps()

		projectID := newTestProjectID(t, "project-1")

		channel := newTestProjectChannel(
			t,
			"channel-1",
			projectID,
			"Project Channel",
			"member-1",
			"member-2",
		)

		deps.channelRepo.channels = []*message.Channel{channel}
		deps.idGen.ids = []string{"message-1"}

		deps.externalBus.publishFn = func(
			ctx context.Context,
			event eventbus.Event,
		) error {
			return errExternalBus
		}

		event := newTestInvoiceStatusChangedEvent(t)
		event.ProjectID = projectID

		subTotal, err := billing.NewDecimal("100.00")
		if err != nil {
			t.Fatal(err)
		}
		event.SubTotal = subTotal

		handler := NewAddInvoiceStatusChangedMessage(
			newTestSystemMessagePublisher(deps),
		)

		err = handler.Handle(context.Background(), event)
		if !errors.Is(err, errExternalBus) {
			t.Fatalf(
				"Handle() error = %v, want %v",
				err,
				errExternalBus,
			)
		}

		if deps.channelRepo.listByProjectCalls != 1 {
			t.Fatalf(
				"ListByProject() calls = %d, want 1",
				deps.channelRepo.listByProjectCalls,
			)
		}

		if deps.messageRepo.addCalls != 1 {
			t.Fatalf(
				"MessageRepository.Add() calls = %d, want 1",
				deps.messageRepo.addCalls,
			)
		}

		if deps.externalBus.publishCalls != 1 {
			t.Fatalf(
				"ExternalBus.Publish() calls = %d, want 1",
				deps.externalBus.publishCalls,
			)
		}
	})
}
