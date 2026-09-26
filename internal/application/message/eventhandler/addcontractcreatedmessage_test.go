package eventhandler

import (
	"context"
	"errors"
	"testing"
	"time"

	"mizu/internal/application/eventbus"
	"mizu/internal/domain/contract"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
)

func TestAddContractCreatedMessage_Handle(t *testing.T) {
	t.Run("ignores unsupported event", func(t *testing.T) {
		deps := newTestDeps()

		publisher := newTestSystemMessagePublisher(deps)
		handler := NewAddContractCreatedMessage(publisher)

		err := handler.Handle(context.Background(), newTestProjectCreatedEvent(t))

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

	t.Run("publishes contract created system message", func(t *testing.T) {
		deps := newTestDeps()

		channelID, err := message.NewChannelID("channel-1")
		if err != nil {
			t.Fatal(err)
		}

		projectID, err := project.NewProjectID("project-1")
		if err != nil {
			t.Fatal(err)
		}

		member1, err := newTestUserID("member-1")
		if err != nil {
			t.Fatal(err)
		}

		member2, err := newTestUserID("member-2")
		if err != nil {
			t.Fatal(err)
		}

		channelName, err := message.NewChannelName("Project Channel")
		if err != nil {
			t.Fatal(err)
		}

		now := time.Now()

		channel, err := message.NewProjectChannel(
			channelID,
			projectID,
			channelName,
			[]iamUserID{member1, member2},
			now,
		)
		if err != nil {
			t.Fatal(err)
		}

		deps.channelRepo.channels = []*message.Channel{channel}

		deps.idGen.ids = []string{"message-1"}

		contractID, err := contract.NewContractID("contract-1")
		if err != nil {
			t.Fatal(err)
		}

		contractName, err := contract.NewName("Website Development")
		if err != nil {
			t.Fatal(err)
		}

		event := contract.ContractCreatedEvent{
			ContractID: contractID,
			ProjectID:  projectID,
			Name:       contractName,
			OccurredAt: now,
		}

		handler := NewAddContractCreatedMessage(
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

		if deps.channelRepo.lastListByProjectID != projectID {
			t.Fatalf(
				"ListByProject() project ID = %q, want %q",
				deps.channelRepo.lastListByProjectID,
				projectID,
			)
		}

		if deps.idGen.generateCalls != 1 {
			t.Fatalf(
				"Generate() calls = %d, want 1",
				deps.idGen.generateCalls,
			)
		}

		if deps.messageRepo.addCalls != 1 {
			t.Fatalf(
				"MessageRepository.Add() calls = %d, want 1",
				deps.messageRepo.addCalls,
			)
		}

		if deps.messageRepo.lastAdded == nil {
			t.Fatal("MessageRepository.Add() received nil message")
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

		broadcast := deps.externalBus.events[0]
		if broadcast == nil {
			t.Fatal("published event is nil")
		}

		if deps.uow.executeCalls != 1 {
			t.Fatalf(
				"UnitOfWork.Execute() calls = %d, want 1",
				deps.uow.executeCalls,
			)
		}

		if deps.uow.callbackCalls != 1 {
			t.Fatalf(
				"UnitOfWork callback calls = %d, want 1",
				deps.uow.callbackCalls,
			)
		}

		// The concrete broadcast payload is checked through its event type here.
		// Detailed payload assertions belong to SystemMessagePublisher tests.
		if broadcast.EventType() == "" {
			t.Fatal("published event has empty event type")
		}
	})

	t.Run("propagates publisher error", func(t *testing.T) {
		deps := newTestDeps()

		expectedErr := errors.New("publish failed")

		deps.externalBus.publishFn = func(
			ctx context.Context,
			event eventbus.Event,
		) error {
			return expectedErr
		}

		projectID, err := project.NewProjectID("project-1")
		if err != nil {
			t.Fatal(err)
		}

		channelID, err := message.NewChannelID("channel-1")
		if err != nil {
			t.Fatal(err)
		}

		member1, err := newTestUserID("member-1")
		if err != nil {
			t.Fatal(err)
		}

		member2, err := newTestUserID("member-2")
		if err != nil {
			t.Fatal(err)
		}

		channelName, err := message.NewChannelName("Project Channel")
		if err != nil {
			t.Fatal(err)
		}

		now := time.Now()

		channel, err := message.NewProjectChannel(
			channelID,
			projectID,
			channelName,
			[]iamUserID{member1, member2},
			now,
		)
		if err != nil {
			t.Fatal(err)
		}

		deps.channelRepo.channels = []*message.Channel{channel}
		deps.idGen.ids = []string{"message-1"}

		contractID, err := contract.NewContractID("contract-1")
		if err != nil {
			t.Fatal(err)
		}

		contractName, err := contract.NewName("Website Development")
		if err != nil {
			t.Fatal(err)
		}

		event := contract.ContractCreatedEvent{
			ContractID: contractID,
			ProjectID:  projectID,
			Name:       contractName,
			OccurredAt: now,
		}

		handler := NewAddContractCreatedMessage(
			newTestSystemMessagePublisher(deps),
		)

		err = handler.Handle(context.Background(), event)
		if !errors.Is(err, expectedErr) {
			t.Fatalf(
				"Handle() error = %v, want %v",
				err,
				expectedErr,
			)
		}
	})
}

// iamUserID is an alias used only to keep the channel fixture concise.
type iamUserID = iam.UserID

func newTestUserID(id string) (iam.UserID, error) {
	return iam.NewUserID(id)
}
