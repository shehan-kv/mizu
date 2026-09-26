package eventhandler

import (
	"context"
	"errors"
	"testing"

	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
)

func TestCreateProjectChannel_Handle(t *testing.T) {
	t.Run("ignores unsupported event", func(t *testing.T) {
		deps := newTestDeps()

		handler := NewCreateProjectChannel(
			deps.channelRepo,
			deps.idGen,
			deps.uow,
		)

		event := newTestUserCreatedEvent(t)

		err := handler.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("Handle() error = %v, want nil", err)
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

		if deps.channelRepo.addCalls != 0 {
			t.Fatalf(
				"ChannelRepository.Add() calls = %d, want 0",
				deps.channelRepo.addCalls,
			)
		}
	})

	t.Run("returns ID generator error", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.generateFn = func() (string, error) {
			return "", errIDGenerator
		}

		handler := NewCreateProjectChannel(
			deps.channelRepo,
			deps.idGen,
			deps.uow,
		)

		event := newTestProjectCreatedEvent(t)

		err := handler.Handle(context.Background(), event)
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

		if deps.uow.executeCalls != 0 {
			t.Fatalf(
				"UnitOfWork.Execute() calls = %d, want 0",
				deps.uow.executeCalls,
			)
		}

		if deps.channelRepo.addCalls != 0 {
			t.Fatalf(
				"ChannelRepository.Add() calls = %d, want 0",
				deps.channelRepo.addCalls,
			)
		}
	})

	t.Run("returns invalid channel ID error", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{""}

		handler := NewCreateProjectChannel(
			deps.channelRepo,
			deps.idGen,
			deps.uow,
		)

		event := newTestProjectCreatedEvent(t)

		err := handler.Handle(context.Background(), event)
		if err == nil {
			t.Fatal("Handle() error = nil, want error")
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

		if deps.channelRepo.addCalls != 0 {
			t.Fatalf(
				"ChannelRepository.Add() calls = %d, want 0",
				deps.channelRepo.addCalls,
			)
		}
	})

	t.Run("returns channel name error", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{"channel-1"}

		event := newTestProjectCreatedEvent(t)
		event.Name, _ = project.NewName("")

		handler := NewCreateProjectChannel(
			deps.channelRepo,
			deps.idGen,
			deps.uow,
		)

		err := handler.Handle(context.Background(), event)
		if err == nil {
			t.Fatal("Handle() error = nil, want error")
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

		if deps.channelRepo.addCalls != 0 {
			t.Fatalf(
				"ChannelRepository.Add() calls = %d, want 0",
				deps.channelRepo.addCalls,
			)
		}
	})

	t.Run("returns channel creation error", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{"channel-1"}

		event := newTestProjectCreatedEvent(t)

		// NewChannel requires at least two members.
		event.Members = nil

		handler := NewCreateProjectChannel(
			deps.channelRepo,
			deps.idGen,
			deps.uow,
		)

		err := handler.Handle(context.Background(), event)
		if err == nil {
			t.Fatal("Handle() error = nil, want error")
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

		if deps.channelRepo.addCalls != 0 {
			t.Fatalf(
				"ChannelRepository.Add() calls = %d, want 0",
				deps.channelRepo.addCalls,
			)
		}
	})

	t.Run("returns unit of work error", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{"channel-1"}

		deps.uow.executeFn = func(
			ctx context.Context,
			fn func(context.Context) error,
		) error {
			return errUOW
		}

		event := newTestProjectCreatedEvent(t)

		name, err := project.NewName("Project 1")
		if err != nil {
			t.Fatal(err)
		}
		event.Name = name

		member1, err := iam.NewUserID("member-1")
		if err != nil {
			t.Fatal(err)
		}

		member2, err := iam.NewUserID("member-2")
		if err != nil {
			t.Fatal(err)
		}

		event.Members = []iam.UserID{member1, member2}

		handler := NewCreateProjectChannel(
			deps.channelRepo,
			deps.idGen,
			deps.uow,
		)

		err = handler.Handle(context.Background(), event)
		if !errors.Is(err, errUOW) {
			t.Fatalf(
				"Handle() error = %v, want %v",
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

		if deps.channelRepo.addCalls != 0 {
			t.Fatalf(
				"ChannelRepository.Add() calls = %d, want 0",
				deps.channelRepo.addCalls,
			)
		}
	})

	t.Run("returns channel repository error", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{"channel-1"}

		deps.channelRepo.addFn = func(
			ctx context.Context,
			channel *message.Channel,
		) error {
			return errChannelRepo
		}

		event := newTestProjectCreatedEvent(t)

		name, err := project.NewName("Project 1")
		if err != nil {
			t.Fatal(err)
		}
		event.Name = name

		member1, err := iam.NewUserID("member-1")
		if err != nil {
			t.Fatal(err)
		}

		member2, err := iam.NewUserID("member-2")
		if err != nil {
			t.Fatal(err)
		}

		event.Members = []iam.UserID{member1, member2}

		handler := NewCreateProjectChannel(
			deps.channelRepo,
			deps.idGen,
			deps.uow,
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

		if deps.channelRepo.addCalls != 1 {
			t.Fatalf(
				"ChannelRepository.Add() calls = %d, want 1",
				deps.channelRepo.addCalls,
			)
		}
	})

	t.Run("creates and adds project channel", func(t *testing.T) {
		deps := newTestDeps()

		deps.idGen.ids = []string{"channel-1"}

		event := newTestProjectCreatedEvent(t)

		name, err := project.NewName("Project 1")
		if err != nil {
			t.Fatal(err)
		}
		event.Name = name

		member1, err := iam.NewUserID("member-1")
		if err != nil {
			t.Fatal(err)
		}

		member2, err := iam.NewUserID("member-2")
		if err != nil {
			t.Fatal(err)
		}

		event.Members = []iam.UserID{member1, member2}

		var addedChannel *message.Channel

		deps.channelRepo.addFn = func(
			ctx context.Context,
			channel *message.Channel,
		) error {
			addedChannel = channel
			return nil
		}

		handler := NewCreateProjectChannel(
			deps.channelRepo,
			deps.idGen,
			deps.uow,
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

		if deps.channelRepo.addCalls != 1 {
			t.Fatalf(
				"ChannelRepository.Add() calls = %d, want 1",
				deps.channelRepo.addCalls,
			)
		}

		if addedChannel == nil {
			t.Fatal("added channel = nil, want channel")
		}

		if addedChannel.ID().String() != "channel-1" {
			t.Fatalf(
				"channel ID = %q, want %q",
				addedChannel.ID().String(),
				"channel-1",
			)
		}

		if addedChannel.ProjectID() == nil {
			t.Fatal("channel ProjectID = nil, want project ID")
		}

		if addedChannel.ProjectID().String() != event.ProjectID.String() {
			t.Fatalf(
				"channel ProjectID = %q, want %q",
				addedChannel.ProjectID().String(),
				event.ProjectID.String(),
			)
		}

		if addedChannel.Name().String() != event.Name.String() {
			t.Fatalf(
				"channel name = %q, want %q",
				addedChannel.Name().String(),
				event.Name.String(),
			)
		}

		members := addedChannel.Members()

		if len(members) != len(event.Members) {
			t.Fatalf(
				"channel members = %d, want %d",
				len(members),
				len(event.Members),
			)
		}

		for i, member := range event.Members {
			if members[i] != member {
				t.Fatalf(
					"members[%d] = %q, want %q",
					i,
					members[i].String(),
					member.String(),
				)
			}
		}

		if !addedChannel.CreatedAt().Equal(event.OccurredAt) {
			t.Fatalf(
				"CreatedAt() = %v, want %v",
				addedChannel.CreatedAt(),
				event.OccurredAt,
			)
		}

		if !addedChannel.UpdatedAt().Equal(event.OccurredAt) {
			t.Fatalf(
				"UpdatedAt() = %v, want %v",
				addedChannel.UpdatedAt(),
				event.OccurredAt,
			)
		}
	})

}
