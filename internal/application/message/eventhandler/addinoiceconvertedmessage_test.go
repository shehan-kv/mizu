package eventhandler

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"mizu/internal/application/eventbus"
	"mizu/internal/application/message/integration"
	"mizu/internal/domain/billing"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
)

func TestAddInvoiceConvertedMessage_Handle(t *testing.T) {
	t.Run("ignores unsupported event", func(t *testing.T) {
		deps := newTestDeps()

		publisher := newTestSystemMessagePublisher(deps)
		handler := NewAddInvoiceConvertedMessage(publisher)

		event := newTestUserCreatedEvent(t)

		err := handler.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if deps.channelRepo.listByProjectCalls != 0 {
			t.Fatalf(
				"ListByProject() calls = %d, want 0",
				deps.channelRepo.listByProjectCalls,
			)
		}

		if deps.idGen.generateCalls != 0 {
			t.Fatalf(
				"Generate() calls = %d, want 0",
				deps.idGen.generateCalls,
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

	t.Run("publishes converted quote system message", func(t *testing.T) {
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

		event := newTestInvoiceConvertedEvent(t)
		event.ProjectID = projectID
		event.IsInvoice = false

		subTotal, err := billing.NewDecimal("100.00")
		if err != nil {
			t.Fatal(err)
		}
		event.SubTotal = subTotal

		handler := NewAddInvoiceConvertedMessage(
			newTestSystemMessagePublisher(deps),
		)

		err = handler.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if deps.channelRepo.listByProjectCalls != 1 {
			t.Fatalf(
				"channelRepo.listByProjectCalls = %d, want 1",
				deps.channelRepo.listByProjectCalls,
			)
		}

		if deps.messageRepo.addCalls != 1 {
			t.Fatalf(
				"messageRepo.addCalls = %d, want 1",
				deps.messageRepo.addCalls,
			)
		}

		if deps.idGen.generateCalls != 1 {
			t.Fatalf(
				"idGen.generateCalls = %d, want 1",
				deps.idGen.generateCalls,
			)
		}

		if len(deps.externalBus.events) != 1 {
			t.Fatalf(
				"externalBus.events = %d, want 1",
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

		if published.ChannelID != channel.ID().String() {
			t.Fatalf(
				"ChannelID = %q, want %q",
				published.ChannelID,
				channel.ID().String(),
			)
		}

		if published.MessageID != "message-1" {
			t.Fatalf(
				"MessageID = %q, want %q",
				published.MessageID,
				"message-1",
			)
		}

		if published.SenderID != iam.SystemUserID.String() {
			t.Fatalf(
				"SenderID = %q, want %q",
				published.SenderID,
				iam.SystemUserID.String(),
			)
		}

		if !published.IsSystem {
			t.Fatal("IsSystem = false, want true")
		}

		expectedRecipients := []string{
			"member-1",
			"member-2",
		}

		if !reflect.DeepEqual(published.To, expectedRecipients) {
			t.Fatalf(
				"To = %v, want %v",
				published.To,
				expectedRecipients,
			)
		}

		var payload struct {
			Type         string `json:"type"`
			DocumentType string `json:"documentType"`
			InvoiceID    string `json:"invoiceId"`
			CurrencyCode string `json:"currencyCode"`
			SubTotal     string `json:"subTotal"`
		}

		if err := json.Unmarshal([]byte(published.Content), &payload); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if payload.Type != billing.EventTypeInvoiceConverted.String() {
			t.Fatalf(
				"Type = %q, want %q",
				payload.Type,
				billing.EventTypeInvoiceConverted.String(),
			)
		}

		if payload.DocumentType != "quote" {
			t.Fatalf(
				"DocumentType = %q, want %q",
				payload.DocumentType,
				"quote",
			)
		}

		if payload.InvoiceID != event.InvoiceID.String() {
			t.Fatalf(
				"InvoiceID = %q, want %q",
				payload.InvoiceID,
				event.InvoiceID.String(),
			)
		}

		if payload.CurrencyCode != event.CurrencyCode.String() {
			t.Fatalf(
				"CurrencyCode = %q, want %q",
				payload.CurrencyCode,
				event.CurrencyCode.String(),
			)
		}

		if payload.SubTotal != "100.00" {
			t.Fatalf(
				"SubTotal = %q, want %q",
				payload.SubTotal,
				"100.00",
			)
		}
	})

	t.Run("publishes converted invoice system message", func(t *testing.T) {
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

		event := newTestInvoiceConvertedEvent(t)
		event.ProjectID = projectID
		event.IsInvoice = true

		subTotal, err := billing.NewDecimal("100.00")
		if err != nil {
			t.Fatal(err)
		}
		event.SubTotal = subTotal

		handler := NewAddInvoiceConvertedMessage(
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
			DocumentType string `json:"documentType"`
			InvoiceID    string `json:"invoiceId"`
			CurrencyCode string `json:"currencyCode"`
			SubTotal     string `json:"subTotal"`
		}

		if err := json.Unmarshal([]byte(published.Content), &payload); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
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

		event := newTestInvoiceConvertedEvent(t)
		event.ProjectID = projectID

		subTotal, err := billing.NewDecimal("100.00")
		if err != nil {
			t.Fatal(err)
		}
		event.SubTotal = subTotal

		handler := NewAddInvoiceConvertedMessage(
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

func newTestProjectID(t *testing.T, id string) project.ProjectID {
	t.Helper()

	projectID, err := project.NewProjectID(id)
	if err != nil {
		t.Fatal(err)
	}

	return projectID
}

func newTestProjectChannel(
	t *testing.T,
	channelID string,
	projectID project.ProjectID,
	name string,
	memberIDs ...string,
) *message.Channel {
	t.Helper()

	id, err := message.NewChannelID(channelID)
	if err != nil {
		t.Fatal(err)
	}

	channelName, err := message.NewChannelName(name)
	if err != nil {
		t.Fatal(err)
	}

	members := make([]iam.UserID, 0, len(memberIDs))

	for _, memberID := range memberIDs {
		userID, err := iam.NewUserID(memberID)
		if err != nil {
			t.Fatal(err)
		}

		members = append(members, userID)
	}

	channel, err := message.NewProjectChannel(
		id,
		projectID,
		channelName,
		members,
		time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}

	return channel
}
