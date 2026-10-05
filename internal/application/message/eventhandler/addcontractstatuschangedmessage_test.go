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

func TestAddContractStatusChangedMessage_Handle(t *testing.T) {
	t.Run("ignores unsupported event", func(t *testing.T) {
		deps := newTestDeps()

		handler := NewAddContractStatusChangedMessage(
			deps.userRepo,
			newTestSystemMessagePublisher(deps),
		)

		err := handler.Handle(
			context.Background(),
			newTestProjectCreatedEvent(t),
		)

		if err != nil {
			t.Fatalf("Handle() error = %v, want nil", err)
		}

		if deps.userRepo.getByIDCalls != 0 {
			t.Fatalf(
				"GetByID() calls = %d, want 0",
				deps.userRepo.getByIDCalls,
			)
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

	t.Run("returns user repository error", func(t *testing.T) {
		deps := newTestDeps()

		expectedErr := errors.New("get user failed")

		deps.userRepo.getByIDFn = func(
			ctx context.Context,
			id iam.UserID,
		) (*iam.User, error) {
			return nil, expectedErr
		}

		event := newTestContractStatusChangedEvent(t)

		handler := NewAddContractStatusChangedMessage(
			deps.userRepo,
			newTestSystemMessagePublisher(deps),
		)

		err := handler.Handle(context.Background(), event)

		if !errors.Is(err, expectedErr) {
			t.Fatalf(
				"Handle() error = %v, want %v",
				err,
				expectedErr,
			)
		}

		if deps.userRepo.getByIDCalls != 1 {
			t.Fatalf(
				"GetByID() calls = %d, want 1",
				deps.userRepo.getByIDCalls,
			)
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

	t.Run("publishes contract status changed system message", func(t *testing.T) {
		deps := newTestDeps()

		userID, err := iam.NewUserID("signatory-1")
		if err != nil {
			t.Fatal(err)
		}

		user := newTestUser(t, "signatory-1", iam.RoleClient)

		deps.userRepo.user = user

		deps.userRepo.getByIDFn = func(
			ctx context.Context,
			id iam.UserID,
		) (*iam.User, error) {
			if id != userID {
				t.Fatalf(
					"GetByID() ID = %q, want %q",
					id,
					userID,
				)
			}

			return user, nil
		}

		projectID, err := project.NewProjectID("project-1")
		if err != nil {
			t.Fatal(err)
		}

		channelID, err := message.NewChannelID("channel-1")
		if err != nil {
			t.Fatal(err)
		}

		member1, err := iam.NewUserID("member-1")
		if err != nil {
			t.Fatal(err)
		}

		member2, err := iam.NewUserID("member-2")
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
			[]iam.UserID{member1, member2},
			now,
		)
		if err != nil {
			t.Fatal(err)
		}

		deps.channelRepo.channels = []*message.Channel{channel}
		deps.idGen.ids = []string{"message-1"}

		event := newTestContractStatusChangedEvent(t)
		event.ProjectID = projectID
		event.Signatory = contract.NewSignatory(userID, now)

		handler := NewAddContractStatusChangedMessage(
			deps.userRepo,
			newTestSystemMessagePublisher(deps),
		)

		err = handler.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if deps.userRepo.getByIDCalls != 1 {
			t.Fatalf(
				"GetByID() calls = %d, want 1",
				deps.userRepo.getByIDCalls,
			)
		}

		if deps.userRepo.lastGetByID != userID {
			t.Fatalf(
				"GetByID() ID = %q, want %q",
				deps.userRepo.lastGetByID,
				userID,
			)
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

		published := deps.externalBus.events[0]

		if published == nil {
			t.Fatal("published event is nil")
		}

		if published.EventType() == "" {
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

		user := newTestUser(t, "signatory-1", iam.RoleClient)
		deps.userRepo.user = user

		projectID, err := project.NewProjectID("project-1")
		if err != nil {
			t.Fatal(err)
		}

		channelID, err := message.NewChannelID("channel-1")
		if err != nil {
			t.Fatal(err)
		}

		member1, err := iam.NewUserID("member-1")
		if err != nil {
			t.Fatal(err)
		}

		member2, err := iam.NewUserID("member-2")
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
			[]iam.UserID{member1, member2},
			now,
		)
		if err != nil {
			t.Fatal(err)
		}

		deps.channelRepo.channels = []*message.Channel{channel}
		deps.idGen.ids = []string{"message-1"}

		event := newTestContractStatusChangedEvent(t)
		event.ProjectID = projectID

		handler := NewAddContractStatusChangedMessage(
			deps.userRepo,
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
