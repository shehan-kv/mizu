package eventhandler

import (
	"context"
	"errors"
	"testing"
	"time"

	"mizu/internal/application/eventbus"
	"mizu/internal/application/message/integration"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
)

func TestSystemMessagePublisher_CreateAndPublish(t *testing.T) {
	t.Run("returns channel repository error", func(t *testing.T) {
		deps := newTestDeps()

		deps.channelRepo.listByProjectFn = func(
			ctx context.Context,
			projectID project.ProjectID,
		) ([]*message.Channel, error) {
			return nil, errChannelRepo
		}

		publisher := newTestSystemMessagePublisher(deps)

		projectID, err := project.NewProjectID("project-1")
		if err != nil {
			t.Fatal(err)
		}

		content, err := message.NewContent("test message")
		if err != nil {
			t.Fatal(err)
		}

		err = publisher.CreateAndPublish(
			context.Background(),
			projectID,
			content,
			time.Now(),
		)

		if !errors.Is(err, errChannelRepo) {
			t.Fatalf(
				"CreateAndPublish() error = %v, want %v",
				err,
				errChannelRepo,
			)
		}

		if deps.channelRepo.listByProjectCalls != 1 {
			t.Fatalf(
				"ChannelRepository.ListByProject() calls = %d, want 1",
				deps.channelRepo.listByProjectCalls,
			)
		}

		if deps.idGen.generateCalls != 0 {
			t.Fatalf(
				"IDGenerator.Generate() calls = %d, want 0",
				deps.idGen.generateCalls,
			)
		}

		if deps.uow.executeCalls != 0 {
			t.Fatalf(
				"UnitOfWork.Execute() calls = %d, want 0",
				deps.uow.executeCalls,
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

	t.Run("does nothing when project has no channels", func(t *testing.T) {
		deps := newTestDeps()

		deps.channelRepo.channels = nil

		publisher := newTestSystemMessagePublisher(deps)

		projectID, err := project.NewProjectID("project-1")
		if err != nil {
			t.Fatal(err)
		}

		content, err := message.NewContent("test message")
		if err != nil {
			t.Fatal(err)
		}

		err = publisher.CreateAndPublish(
			context.Background(),
			projectID,
			content,
			time.Now(),
		)
		if err != nil {
			t.Fatalf("CreateAndPublish() error = %v, want nil", err)
		}

		if deps.channelRepo.listByProjectCalls != 1 {
			t.Fatalf(
				"ChannelRepository.ListByProject() calls = %d, want 1",
				deps.channelRepo.listByProjectCalls,
			)
		}

		if deps.idGen.generateCalls != 0 {
			t.Fatalf(
				"IDGenerator.Generate() calls = %d, want 0",
				deps.idGen.generateCalls,
			)
		}

		if deps.uow.executeCalls != 0 {
			t.Fatalf(
				"UnitOfWork.Execute() calls = %d, want 0",
				deps.uow.executeCalls,
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

	t.Run("returns ID generator error", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.generateFn = func() (string, error) {
			return "", errIDGenerator
		}

		projectID, err := project.NewProjectID("project-1")
		if err != nil {
			t.Fatal(err)
		}

		now := time.Now()

		channel := newTestProjectChannel(
			t,
			"channel-1",
			projectID,
			"General",
			"member-1",
			"member-2",
		)

		deps.channelRepo.channels = []*message.Channel{channel}

		publisher := newTestSystemMessagePublisher(deps)

		content, err := message.NewContent("test message")
		if err != nil {
			t.Fatal(err)
		}

		err = publisher.CreateAndPublish(
			context.Background(),
			projectID,
			content,
			now,
		)

		if !errors.Is(err, errIDGenerator) {
			t.Fatalf(
				"CreateAndPublish() error = %v, want %v",
				err,
				errIDGenerator,
			)
		}

		if deps.idGen.generateCalls != 1 {
			t.Fatalf(
				"IDGenerator.Generate() calls = %d, want 1",
				deps.idGen.generateCalls,
			)
		}

		if deps.uow.executeCalls != 0 {
			t.Fatalf(
				"UnitOfWork.Execute() calls = %d, want 0",
				deps.uow.executeCalls,
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

	t.Run("returns invalid message ID error", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{""}

		projectID, err := project.NewProjectID("project-1")
		if err != nil {
			t.Fatal(err)
		}

		channel := newTestProjectChannel(
			t,
			"channel-1",
			projectID,
			"General",
			"member-1",
			"member-2",
		)

		deps.channelRepo.channels = []*message.Channel{channel}

		publisher := newTestSystemMessagePublisher(deps)

		content, err := message.NewContent("test message")
		if err != nil {
			t.Fatal(err)
		}

		err = publisher.CreateAndPublish(
			context.Background(),
			projectID,
			content,
			time.Now(),
		)

		if err == nil {
			t.Fatal("CreateAndPublish() error = nil, want error")
		}

		if deps.idGen.generateCalls != 1 {
			t.Fatalf(
				"IDGenerator.Generate() calls = %d, want 1",
				deps.idGen.generateCalls,
			)
		}

		if deps.uow.executeCalls != 0 {
			t.Fatalf(
				"UnitOfWork.Execute() calls = %d, want 0",
				deps.uow.executeCalls,
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

	t.Run("returns unit of work error", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{"message-1"}

		deps.uow.executeFn = func(
			ctx context.Context,
			fn func(context.Context) error,
		) error {
			return errUOW
		}

		projectID, err := project.NewProjectID("project-1")
		if err != nil {
			t.Fatal(err)
		}

		channel := newTestProjectChannel(
			t,
			"channel-1",
			projectID,
			"General",
			"member-1",
			"member-2",
		)

		deps.channelRepo.channels = []*message.Channel{channel}

		publisher := newTestSystemMessagePublisher(deps)

		content, err := message.NewContent("test message")
		if err != nil {
			t.Fatal(err)
		}

		err = publisher.CreateAndPublish(
			context.Background(),
			projectID,
			content,
			time.Now(),
		)

		if !errors.Is(err, errUOW) {
			t.Fatalf(
				"CreateAndPublish() error = %v, want %v",
				err,
				errUOW,
			)
		}

		if deps.idGen.generateCalls != 1 {
			t.Fatalf(
				"IDGenerator.Generate() calls = %d, want 1",
				deps.idGen.generateCalls,
			)
		}

		if deps.uow.executeCalls != 1 {
			t.Fatalf(
				"UnitOfWork.Execute() calls = %d, want 1",
				deps.uow.executeCalls,
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

	t.Run("returns message repository error", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{"message-1"}

		deps.messageRepo.addFn = func(
			ctx context.Context,
			msg *message.Message,
		) error {
			return errMessageRepo
		}

		projectID, err := project.NewProjectID("project-1")
		if err != nil {
			t.Fatal(err)
		}

		channel := newTestProjectChannel(
			t,
			"channel-1",
			projectID,
			"General",
			"member-1",
			"member-2",
		)

		deps.channelRepo.channels = []*message.Channel{channel}

		publisher := newTestSystemMessagePublisher(deps)

		content, err := message.NewContent("test message")
		if err != nil {
			t.Fatal(err)
		}

		err = publisher.CreateAndPublish(
			context.Background(),
			projectID,
			content,
			time.Now(),
		)

		if !errors.Is(err, errMessageRepo) {
			t.Fatalf(
				"CreateAndPublish() error = %v, want %v",
				err,
				errMessageRepo,
			)
		}

		if deps.idGen.generateCalls != 1 {
			t.Fatalf(
				"IDGenerator.Generate() calls = %d, want 1",
				deps.idGen.generateCalls,
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

		if deps.messageRepo.addCalls != 1 {
			t.Fatalf(
				"MessageRepository.Add() calls = %d, want 1",
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

	t.Run("creates persists and publishes system messages", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{
			"message-1",
			"message-2",
		}

		var addedMessages []*message.Message
		deps.messageRepo.addFn = func(
			ctx context.Context,
			msg *message.Message,
		) error {
			addedMessages = append(addedMessages, msg)
			return nil
		}

		projectID, err := project.NewProjectID("project-1")
		if err != nil {
			t.Fatal(err)
		}

		now := time.Now()

		channel1 := newTestProjectChannel(
			t,
			"channel-1",
			projectID,
			"General",
			"member-1",
			"member-2",
		)

		channel2 := newTestProjectChannel(
			t,
			"channel-2",
			projectID,
			"Development",
			"member-2",
			"member-3",
		)

		deps.channelRepo.channels = []*message.Channel{
			channel1,
			channel2,
		}

		publisher := newTestSystemMessagePublisher(deps)

		content, err := message.NewContent("Project updated")
		if err != nil {
			t.Fatal(err)
		}

		err = publisher.CreateAndPublish(
			context.Background(),
			projectID,
			content,
			now,
		)
		if err != nil {
			t.Fatalf("CreateAndPublish() error = %v", err)
		}

		if deps.channelRepo.listByProjectCalls != 1 {
			t.Fatalf(
				"ChannelRepository.ListByProject() calls = %d, want 1",
				deps.channelRepo.listByProjectCalls,
			)
		}

		if deps.idGen.generateCalls != 2 {
			t.Fatalf(
				"IDGenerator.Generate() calls = %d, want 2",
				deps.idGen.generateCalls,
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

		if deps.messageRepo.addCalls != 2 {
			t.Fatalf(
				"MessageRepository.Add() calls = %d, want 2",
				deps.messageRepo.addCalls,
			)
		}

		if len(addedMessages) != 2 {
			t.Fatalf(
				"persisted messages = %d, want 2",
				len(addedMessages),
			)
		}

		expectedIDs := []string{
			"message-1",
			"message-2",
		}

		for i, msg := range addedMessages {
			if msg == nil {
				t.Fatalf("persisted message[%d] = nil", i)
			}

			if msg.ID().String() != expectedIDs[i] {
				t.Fatalf(
					"message[%d] ID = %q, want %q",
					i,
					msg.ID().String(),
					expectedIDs[i],
				)
			}

			if !msg.CreatedAt().Equal(now) {
				t.Fatalf(
					"message[%d] CreatedAt() = %v, want %v",
					i,
					msg.CreatedAt(),
					now,
				)
			}
		}

		if deps.externalBus.publishCalls != 2 {
			t.Fatalf(
				"ExternalBus.Publish() calls = %d, want 2",
				deps.externalBus.publishCalls,
			)
		}

		if len(deps.externalBus.events) != 2 {
			t.Fatalf(
				"published events = %d, want 2",
				len(deps.externalBus.events),
			)
		}

		expectedContent := content.String()

		expectedEvents := map[string]struct {
			channelID string
			members   []iam.UserID
		}{
			"message-1": {
				channelID: channel1.ID().String(),
				members:   channel1.Members(),
			},
			"message-2": {
				channelID: channel2.ID().String(),
				members:   channel2.Members(),
			},
		}

		for i, rawEvent := range deps.externalBus.events {
			published, ok := rawEvent.(integration.MessageBroadcast)
			if !ok {
				t.Fatalf(
					"event[%d] type = %T, want integration.MessageBroadcast",
					i,
					rawEvent,
				)
			}

			expected, ok := expectedEvents[published.MessageID]
			if !ok {
				t.Fatalf(
					"event[%d] MessageID = %q, want message-1 or message-2",
					i,
					published.MessageID,
				)
			}

			if published.ChannelID != expected.channelID {
				t.Fatalf(
					"event[%d] ChannelID = %q, want %q",
					i,
					published.ChannelID,
					expected.channelID,
				)
			}

			if published.SenderID != iam.SystemUserID.String() {
				t.Fatalf(
					"event[%d] SenderID = %q, want %q",
					i,
					published.SenderID,
					iam.SystemUserID.String(),
				)
			}

			if !published.IsSystem {
				t.Fatalf("event[%d] IsSystem = false, want true", i)
			}

			if published.Content != expectedContent {
				t.Fatalf(
					"event[%d] Content = %q, want %q",
					i,
					published.Content,
					expectedContent,
				)
			}

			if !published.OccurredAt.Equal(now) {
				t.Fatalf(
					"event[%d] OccurredAt() = %v, want %v",
					i,
					published.OccurredAt,
					now,
				)
			}

			if len(published.To) != len(expected.members) {
				t.Fatalf(
					"event[%d] To length = %d, want %d",
					i,
					len(published.To),
					len(expected.members),
				)
			}

			for j, member := range expected.members {
				if published.To[j] != member.String() {
					t.Fatalf(
						"event[%d] To[%d] = %q, want %q",
						i,
						j,
						published.To[j],
						member.String(),
					)
				}
			}
		}
	})

	t.Run("returns publishing error", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{"message-1"}

		deps.externalBus.publishFn = func(
			ctx context.Context,
			event eventbus.Event,
		) error {
			return errExternalBus
		}

		projectID, err := project.NewProjectID("project-1")
		if err != nil {
			t.Fatal(err)
		}

		channel := newTestProjectChannel(
			t,
			"channel-1",
			projectID,
			"General",
			"member-1",
			"member-2",
		)

		deps.channelRepo.channels = []*message.Channel{channel}

		publisher := newTestSystemMessagePublisher(deps)

		content, err := message.NewContent("test message")
		if err != nil {
			t.Fatal(err)
		}

		err = publisher.CreateAndPublish(
			context.Background(),
			projectID,
			content,
			time.Now(),
		)

		if !errors.Is(err, errExternalBus) {
			t.Fatalf(
				"CreateAndPublish() error = %v, want %v",
				err,
				errExternalBus,
			)
		}

		if deps.idGen.generateCalls != 1 {
			t.Fatalf(
				"IDGenerator.Generate() calls = %d, want 1",
				deps.idGen.generateCalls,
			)
		}

		if deps.uow.executeCalls != 1 {
			t.Fatalf(
				"UnitOfWork.Execute() calls = %d, want 1",
				deps.uow.executeCalls,
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

	t.Run("publishes all events and joins multiple publishing errors", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{
			"message-1",
			"message-2",
			"message-3",
		}

		errPublish1 := errors.New("publish error 1")
		errPublish2 := errors.New("publish error 2")

		deps.externalBus.publishFn = func(
			ctx context.Context,
			event eventbus.Event,
		) error {
			published, ok := event.(integration.MessageBroadcast)
			if !ok {
				return errors.New("unexpected event type")
			}

			switch published.MessageID {
			case "message-1":
				return errPublish1
			case "message-2":
				return nil
			case "message-3":
				return errPublish2
			default:
				return nil
			}
		}

		projectID, err := project.NewProjectID("project-1")
		if err != nil {
			t.Fatal(err)
		}

		deps.channelRepo.channels = []*message.Channel{
			newTestProjectChannel(
				t,
				"channel-1",
				projectID,
				"General",
				"member-1",
				"member-2",
			),
			newTestProjectChannel(
				t,
				"channel-2",
				projectID,
				"Development",
				"member-2",
				"member-3",
			),
			newTestProjectChannel(
				t,
				"channel-3",
				projectID,
				"Support",
				"member-3",
				"member-4",
			),
		}

		publisher := newTestSystemMessagePublisher(deps)

		content, err := message.NewContent("test message")
		if err != nil {
			t.Fatal(err)
		}

		err = publisher.CreateAndPublish(
			context.Background(),
			projectID,
			content,
			time.Now(),
		)

		if err == nil {
			t.Fatal("CreateAndPublish() error = nil, want error")
		}

		if !errors.Is(err, errPublish1) {
			t.Fatalf(
				"CreateAndPublish() error does not contain errPublish1: %v",
				err,
			)
		}

		if !errors.Is(err, errPublish2) {
			t.Fatalf(
				"CreateAndPublish() error does not contain errPublish2: %v",
				err,
			)
		}

		if deps.idGen.generateCalls != 3 {
			t.Fatalf(
				"IDGenerator.Generate() calls = %d, want 3",
				deps.idGen.generateCalls,
			)
		}

		if deps.uow.executeCalls != 1 {
			t.Fatalf(
				"UnitOfWork.Execute() calls = %d, want 1",
				deps.uow.executeCalls,
			)
		}

		if deps.messageRepo.addCalls != 3 {
			t.Fatalf(
				"MessageRepository.Add() calls = %d, want 3",
				deps.messageRepo.addCalls,
			)
		}

		if deps.externalBus.publishCalls != 3 {
			t.Fatalf(
				"ExternalBus.Publish() calls = %d, want 3",
				deps.externalBus.publishCalls,
			)
		}
	})
}
