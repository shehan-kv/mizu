package eventhandler

import (
	"context"
	"errors"
	"testing"

	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
)

func TestCreateDefaultChannel_Handle(t *testing.T) {
	t.Run("ignores unsupported event", func(t *testing.T) {
		deps := newTestDeps()

		handler := NewCreateDefaultChannel(
			nil,
			deps.channelRepo,
			deps.idGen,
		)

		err := handler.Handle(
			context.Background(),
			newTestProjectCreatedEvent(t),
		)
		if err != nil {
			t.Fatalf("Handle() error = %v, want nil", err)
		}

		if deps.idGen.generateCalls != 0 {
			t.Fatalf(
				"IDGenerator.Generate() calls = %d, want 0",
				deps.idGen.generateCalls,
			)
		}

		if deps.channelRepo.saveCalls != 0 {
			t.Fatalf(
				"ChannelRepository.Save() calls = %d, want 0",
				deps.channelRepo.saveCalls,
			)
		}
	})

	t.Run("returns ID generator error", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.generateFn = func() (string, error) {
			return "", errIDGenerator
		}

		handler := NewCreateDefaultChannel(
			nil,
			deps.channelRepo,
			deps.idGen,
		)

		err := handler.Handle(
			context.Background(),
			newTestUserCreatedEvent(t),
		)
		if !errors.Is(err, errIDGenerator) {
			t.Fatalf(
				"Handle() error = %v, want %v",
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

		if deps.channelRepo.saveCalls != 0 {
			t.Fatalf(
				"ChannelRepository.Save() calls = %d, want 0",
				deps.channelRepo.saveCalls,
			)
		}
	})

	t.Run("returns channel repository error", func(t *testing.T) {
		deps := newTestDeps()

		deps.channelRepo.saveFn = func(
			ctx context.Context,
			channel *message.Channel,
		) error {
			return errChannelRepo
		}

		event := newTestUserCreatedEvent(t)

		createdBy, err := iam.NewUserID("creator-1")
		if err != nil {
			t.Fatal(err)
		}
		event.CreatedBy = createdBy

		handler := NewCreateDefaultChannel(
			nil,
			deps.channelRepo,
			deps.idGen,
		)

		err = handler.Handle(context.Background(), event)
		if !errors.Is(err, errChannelRepo) {
			t.Fatalf(
				"Handle() error = %v, want %v",
				err,
				errChannelRepo,
			)
		}

		if deps.idGen.generateCalls != 1 {
			t.Fatalf(
				"IDGenerator.Generate() calls = %d, want 1",
				deps.idGen.generateCalls,
			)
		}

		if deps.channelRepo.saveCalls != 1 {
			t.Fatalf(
				"ChannelRepository.Save() calls = %d, want 1",
				deps.channelRepo.saveCalls,
			)
		}
	})

	t.Run("creates and saves default channel", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{"channel-1"}

		event := newTestUserCreatedEvent(t)

		createdBy, err := iam.NewUserID("creator-1")
		if err != nil {
			t.Fatal(err)
		}
		event.CreatedBy = createdBy

		var savedChannel *message.Channel

		deps.channelRepo.saveFn = func(
			ctx context.Context,
			channel *message.Channel,
		) error {
			savedChannel = channel
			return nil
		}

		handler := NewCreateDefaultChannel(
			nil,
			deps.channelRepo,
			deps.idGen,
		)

		err = handler.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if deps.idGen.generateCalls != 1 {
			t.Fatalf(
				"IDGenerator.Generate() calls = %d, want 1",
				deps.idGen.generateCalls,
			)
		}

		if deps.channelRepo.saveCalls != 1 {
			t.Fatalf(
				"ChannelRepository.Save() calls = %d, want 1",
				deps.channelRepo.saveCalls,
			)
		}

		if savedChannel == nil {
			t.Fatal("saved channel = nil, want channel")
		}

		channel := savedChannel

		if channel.ID().String() != "channel-1" {
			t.Fatalf(
				"channel ID = %q, want %q",
				channel.ID().String(),
				"channel-1",
			)
		}

		if channel.ProjectID() != nil {
			t.Fatalf("channel ProjectID = %v, want nil", channel.ProjectID())
		}

		expectedName := event.Name.FirstName() +
			" " +
			event.Name.LastName() +
			" · Workspace"

		if channel.Name().String() != expectedName {
			t.Fatalf(
				"channel name = %q, want %q",
				channel.Name().String(),
				expectedName,
			)
		}

		if !channel.HasMember(event.UserID) {
			t.Fatalf(
				"channel does not contain created user %q",
				event.UserID.String(),
			)
		}

		if !channel.HasMember(event.CreatedBy) {
			t.Fatalf(
				"channel does not contain creator %q",
				event.CreatedBy.String(),
			)
		}

		members := channel.Members()

		if len(members) != 2 {
			t.Fatalf(
				"channel members = %d, want 2",
				len(members),
			)
		}

		if members[0] != event.UserID {
			t.Fatalf(
				"members[0] = %q, want %q",
				members[0].String(),
				event.UserID.String(),
			)
		}

		if members[1] != event.CreatedBy {
			t.Fatalf(
				"members[1] = %q, want %q",
				members[1].String(),
				event.CreatedBy.String(),
			)
		}

		if !channel.CreatedAt().Equal(event.OccurredAt) {
			t.Fatalf(
				"CreatedAt() = %v, want %v",
				channel.CreatedAt(),
				event.OccurredAt,
			)
		}

		if !channel.UpdatedAt().Equal(event.OccurredAt) {
			t.Fatalf(
				"UpdatedAt() = %v, want %v",
				channel.UpdatedAt(),
				event.OccurredAt,
			)
		}
	})

	t.Run("returns error when generated channel ID is invalid", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{""}

		handler := NewCreateDefaultChannel(
			nil,
			deps.channelRepo,
			deps.idGen,
		)

		err := handler.Handle(
			context.Background(),
			newTestUserCreatedEvent(t),
		)
		if err == nil {
			t.Fatal("Handle() error = nil, want error")
		}

		if deps.channelRepo.saveCalls != 0 {
			t.Fatalf(
				"ChannelRepository.Save() calls = %d, want 0",
				deps.channelRepo.saveCalls,
			)
		}
	})

	_ = iam.RoleClient
}
