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
)

func TestAddFileUploadedMessage_Handle(t *testing.T) {
	t.Run("ignores unsupported event", func(t *testing.T) {
		deps := newTestDeps()

		handler := NewAddFileUploadedMessage(
			deps.messageRepo,
			deps.channelRepo,
			deps.userRepo,
			deps.externalBus,
			deps.idGen,
		)

		event := newTestUserCreatedEvent(t)

		err := handler.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if deps.userRepo.getByIDCalls != 0 {
			t.Fatalf(
				"GetByID() calls = %d, want 0",
				deps.userRepo.getByIDCalls,
			)
		}

		if deps.channelRepo.getCalls != 0 {
			t.Fatalf(
				"Get() calls = %d, want 0",
				deps.channelRepo.getCalls,
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

	t.Run("returns user repository error", func(t *testing.T) {
		deps := newTestDeps()

		userID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		deps.userRepo.getByIDFn = func(
			ctx context.Context,
			id iam.UserID,
		) (*iam.User, error) {
			return nil, errUserRepo
		}

		event := newTestFileCreatedEvent(t)
		event.UserID = userID

		handler := NewAddFileUploadedMessage(
			deps.messageRepo,
			deps.channelRepo,
			deps.userRepo,
			deps.externalBus,
			deps.idGen,
		)

		err = handler.Handle(context.Background(), event)
		if !errors.Is(err, errUserRepo) {
			t.Fatalf(
				"Handle() error = %v, want %v",
				err,
				errUserRepo,
			)
		}

		if deps.userRepo.getByIDCalls != 1 {
			t.Fatalf(
				"GetByID() calls = %d, want 1",
				deps.userRepo.getByIDCalls,
			)
		}

		if deps.channelRepo.getCalls != 0 {
			t.Fatalf(
				"Get() calls = %d, want 0",
				deps.channelRepo.getCalls,
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

	t.Run("returns channel repository error", func(t *testing.T) {
		deps := newTestDeps()

		user := newTestUser(t, "user-1", iam.RoleClient)
		deps.userRepo.user = user

		deps.channelRepo.getFn = func(
			ctx context.Context,
			id message.ChannelID,
		) (*message.Channel, error) {
			return nil, errChannelRepo
		}

		event := newTestFileCreatedEvent(t)
		event.UserID = user.ID()

		handler := NewAddFileUploadedMessage(
			deps.messageRepo,
			deps.channelRepo,
			deps.userRepo,
			deps.externalBus,
			deps.idGen,
		)

		err := handler.Handle(context.Background(), event)
		if !errors.Is(err, errChannelRepo) {
			t.Fatalf(
				"Handle() error = %v, want %v",
				err,
				errChannelRepo,
			)
		}

		if deps.userRepo.getByIDCalls != 1 {
			t.Fatalf(
				"GetByID() calls = %d, want 1",
				deps.userRepo.getByIDCalls,
			)
		}

		if deps.channelRepo.getCalls != 1 {
			t.Fatalf(
				"Get() calls = %d, want 1",
				deps.channelRepo.getCalls,
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

	t.Run("returns ID generator error", func(t *testing.T) {
		deps := newTestDeps()

		user := newTestUser(t, "user-1", iam.RoleClient)
		deps.userRepo.user = user

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

		channel, err := message.NewUserChannel(
			channelID,
			channelName,
			[]iam.UserID{member1, member2},
			now,
		)
		if err != nil {
			t.Fatal(err)
		}

		deps.channelRepo.channel = channel

		deps.idGen.generateFn = func() (string, error) {
			return "", errIDGenerator
		}

		event := newTestFileCreatedEvent(t)
		event.UserID = user.ID()
		event.ChannelID = channelID

		handler := NewAddFileUploadedMessage(
			deps.messageRepo,
			deps.channelRepo,
			deps.userRepo,
			deps.externalBus,
			deps.idGen,
		)

		err = handler.Handle(context.Background(), event)
		if !errors.Is(err, errIDGenerator) {
			t.Fatalf(
				"Handle() error = %v, want %v",
				err,
				errIDGenerator,
			)
		}

		if deps.userRepo.getByIDCalls != 1 {
			t.Fatalf(
				"GetByID() calls = %d, want 1",
				deps.userRepo.getByIDCalls,
			)
		}

		if deps.channelRepo.getCalls != 1 {
			t.Fatalf(
				"Get() calls = %d, want 1",
				deps.channelRepo.getCalls,
			)
		}

		if deps.idGen.generateCalls != 1 {
			t.Fatalf(
				"Generate() calls = %d, want 1",
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

	t.Run("returns message repository error", func(t *testing.T) {
		deps := newTestDeps()

		user := newTestUser(t, "user-1", iam.RoleClient)
		deps.userRepo.user = user

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

		channel, err := message.NewUserChannel(
			channelID,
			channelName,
			[]iam.UserID{member1, member2},
			now,
		)
		if err != nil {
			t.Fatal(err)
		}

		deps.channelRepo.channel = channel

		deps.messageRepo.addFn = func(
			ctx context.Context,
			m *message.Message,
		) error {
			return errMessageRepo
		}

		event := newTestFileCreatedEvent(t)
		event.UserID = user.ID()
		event.ChannelID = channelID

		handler := NewAddFileUploadedMessage(
			deps.messageRepo,
			deps.channelRepo,
			deps.userRepo,
			deps.externalBus,
			deps.idGen,
		)

		err = handler.Handle(context.Background(), event)
		if !errors.Is(err, errMessageRepo) {
			t.Fatalf(
				"Handle() error = %v, want %v",
				err,
				errMessageRepo,
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

		if deps.externalBus.publishCalls != 0 {
			t.Fatalf(
				"ExternalBus.Publish() calls = %d, want 0",
				deps.externalBus.publishCalls,
			)
		}
	})

	t.Run("returns external bus error", func(t *testing.T) {
		deps := newTestDeps()

		user := newTestUser(t, "user-1", iam.RoleClient)
		deps.userRepo.user = user

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

		channel, err := message.NewUserChannel(
			channelID,
			channelName,
			[]iam.UserID{member1, member2},
			now,
		)
		if err != nil {
			t.Fatal(err)
		}

		deps.channelRepo.channel = channel

		deps.externalBus.publishFn = func(
			ctx context.Context,
			event eventbus.Event,
		) error {
			return errExternalBus
		}

		event := newTestFileCreatedEvent(t)
		event.UserID = user.ID()
		event.ChannelID = channelID

		handler := NewAddFileUploadedMessage(
			deps.messageRepo,
			deps.channelRepo,
			deps.userRepo,
			deps.externalBus,
			deps.idGen,
		)

		err = handler.Handle(context.Background(), event)
		if !errors.Is(err, errExternalBus) {
			t.Fatalf(
				"Handle() error = %v, want %v",
				err,
				errExternalBus,
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
	})

	t.Run("publishes file uploaded system message", func(t *testing.T) {
		deps := newTestDeps()

		user := newTestUser(t, "user-1", iam.RoleClient)
		deps.userRepo.user = user

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

		channel, err := message.NewUserChannel(
			channelID,
			channelName,
			[]iam.UserID{member1, member2},
			now,
		)
		if err != nil {
			t.Fatal(err)
		}

		deps.channelRepo.channel = channel
		deps.idGen.ids = []string{"message-1"}

		event := newTestFileCreatedEvent(t)
		event.UserID = user.ID()
		event.ChannelID = channelID
		event.OccurredAt = now

		handler := NewAddFileUploadedMessage(
			deps.messageRepo,
			deps.channelRepo,
			deps.userRepo,
			deps.externalBus,
			deps.idGen,
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

		if deps.userRepo.lastGetByID != user.ID() {
			t.Fatalf(
				"GetByID() ID = %q, want %q",
				deps.userRepo.lastGetByID,
				user.ID(),
			)
		}

		if deps.channelRepo.getCalls != 1 {
			t.Fatalf(
				"Get() calls = %d, want 1",
				deps.channelRepo.getCalls,
			)
		}

		if deps.channelRepo.lastGetID != channelID {
			t.Fatalf(
				"Get() channel ID = %q, want %q",
				deps.channelRepo.lastGetID,
				channelID,
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

		if deps.messageRepo.lastAdded.ID().String() != "message-1" {
			t.Fatalf(
				"message ID = %q, want %q",
				deps.messageRepo.lastAdded.ID(),
				"message-1",
			)
		}

		if deps.messageRepo.lastAdded.ChannelID() != channelID {
			t.Fatalf(
				"message channel ID = %q, want %q",
				deps.messageRepo.lastAdded.ChannelID(),
				channelID,
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

		if published.MessageID != "message-1" {
			t.Fatalf(
				"MessageID = %q, want %q",
				published.MessageID,
				"message-1",
			)
		}

		if published.ChannelID != channelID.String() {
			t.Fatalf(
				"ChannelID = %q, want %q",
				published.ChannelID,
				channelID.String(),
			)
		}

		if len(published.To) != 2 {
			t.Fatalf(
				"To length = %d, want 2",
				len(published.To),
			)
		}

		if published.To[0] != member1.String() {
			t.Fatalf(
				"To[0] = %q, want %q",
				published.To[0],
				member1.String(),
			)
		}

		if published.To[1] != member2.String() {
			t.Fatalf(
				"To[1] = %q, want %q",
				published.To[1],
				member2.String(),
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

		if published.Content == "" {
			t.Fatal("Content is empty")
		}

		if published.OccurredAt != now {
			t.Fatalf(
				"OccurredAt = %v, want %v",
				published.OccurredAt,
				now,
			)
		}
	})
}
