package message

import (
	"testing"
	"time"

	"mizu/internal/domain/iam"
)

func TestNewUserMessage(t *testing.T) {
	now := time.Now()

	messageID, err := NewMessageID("message-1")
	if err != nil {
		t.Fatal(err)
	}

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	senderID, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	content, err := NewContent("Hello")
	if err != nil {
		t.Fatal(err)
	}

	message, err := NewUserMessage(
		messageID,
		channelID,
		senderID,
		content,
		now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if message.ID() != messageID {
		t.Fatalf("expected ID %q, got %q", messageID, message.ID())
	}

	if message.ChannelID() != channelID {
		t.Fatalf("expected channel ID %q, got %q", channelID, message.ChannelID())
	}

	if message.SenderID() != senderID {
		t.Fatalf("expected sender ID %q, got %q", senderID, message.SenderID())
	}

	if message.Content() != content {
		t.Fatalf("expected content %q, got %q", content, message.Content())
	}

	if message.CreatedAt() != now {
		t.Fatalf("expected createdAt %v, got %v", now, message.CreatedAt())
	}

	if !message.IsUserMessage() {
		t.Fatal("expected user message")
	}

	if message.IsSystemMessage() {
		t.Fatal("expected message not to be a system message")
	}
}

func TestNewUserMessageRejectsSystemUser(t *testing.T) {
	now := time.Now()

	messageID, err := NewMessageID("message-1")
	if err != nil {
		t.Fatal(err)
	}

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	content, err := NewContent("System message")
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewUserMessage(
		messageID,
		channelID,
		iam.SystemUserID,
		content,
		now,
	)

	if err != ErrMessageInvalidSender {
		t.Fatalf("expected %v, got %v", ErrMessageInvalidSender, err)
	}
}

func TestNewSystemMessage(t *testing.T) {
	now := time.Now()

	messageID, err := NewMessageID("message-1")
	if err != nil {
		t.Fatal(err)
	}

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	content, err := NewContent("System message")
	if err != nil {
		t.Fatal(err)
	}

	message := NewSystemMessage(
		messageID,
		channelID,
		content,
		now,
	)

	if message.ID() != messageID {
		t.Fatalf("expected ID %q, got %q", messageID, message.ID())
	}

	if message.ChannelID() != channelID {
		t.Fatalf("expected channel ID %q, got %q", channelID, message.ChannelID())
	}

	if message.SenderID() != iam.SystemUserID {
		t.Fatalf("expected system sender %q, got %q", iam.SystemUserID, message.SenderID())
	}

	if message.Content() != content {
		t.Fatalf("expected content %q, got %q", content, message.Content())
	}

	if !message.IsSystemMessage() {
		t.Fatal("expected system message")
	}

	if message.IsUserMessage() {
		t.Fatal("expected message not to be a user message")
	}

	if message.CreatedAt() != now {
		t.Fatalf("expected createdAt %v, got %v", now, message.CreatedAt())
	}
}

func TestRestoreMessage(t *testing.T) {
	now := time.Now()

	messageID, err := NewMessageID("message-1")
	if err != nil {
		t.Fatal(err)
	}

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	senderID, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	content, err := NewContent("Restored message")
	if err != nil {
		t.Fatal(err)
	}

	message := RestoreMessage(
		messageID,
		channelID,
		senderID,
		false,
		content,
		now,
	)

	if message.ID() != messageID {
		t.Fatalf("expected ID %q, got %q", messageID, message.ID())
	}

	if message.ChannelID() != channelID {
		t.Fatalf("expected channel ID %q, got %q", channelID, message.ChannelID())
	}

	if message.SenderID() != senderID {
		t.Fatalf("expected sender ID %q, got %q", senderID, message.SenderID())
	}

	if message.Content() != content {
		t.Fatalf("expected content %q, got %q", content, message.Content())
	}

	if message.IsSystemMessage() {
		t.Fatal("expected restored user message")
	}

	if !message.IsUserMessage() {
		t.Fatal("expected restored user message")
	}

	if message.CreatedAt() != now {
		t.Fatalf("expected createdAt %v, got %v", now, message.CreatedAt())
	}
}

func TestMessageEquals(t *testing.T) {
	now := time.Now()

	messageID, err := NewMessageID("message-1")
	if err != nil {
		t.Fatal(err)
	}

	otherID, err := NewMessageID("message-2")
	if err != nil {
		t.Fatal(err)
	}

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	senderID, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	content, err := NewContent("Hello")
	if err != nil {
		t.Fatal(err)
	}

	message1, err := NewUserMessage(
		messageID,
		channelID,
		senderID,
		content,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}

	message2, err := NewUserMessage(
		messageID,
		channelID,
		senderID,
		content,
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}

	message3, err := NewUserMessage(
		otherID,
		channelID,
		senderID,
		content,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !message1.Equals(*message2) {
		t.Fatal("messages with the same ID should be equal")
	}

	if message1.Equals(*message3) {
		t.Fatal("messages with different IDs should not be equal")
	}
}
