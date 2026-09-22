package message

import (
	"testing"
	"time"

	"mizu/internal/domain/iam"
)

func TestNewFile(t *testing.T) {
	now := time.Now()

	fileID, err := NewFileID("file-1")
	if err != nil {
		t.Fatal(err)
	}

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	userID, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	originalName, err := NewFileName("document.pdf")
	if err != nil {
		t.Fatal(err)
	}

	savedName, err := NewFileName("01HXYZ-document.pdf")
	if err != nil {
		t.Fatal(err)
	}

	const (
		storageKey = "channels/channel-1/01HXYZ-document.pdf"
		mimeType   = "application/pdf"
		size       = int64(1024)
	)

	file := NewFile(
		fileID,
		channelID,
		userID,
		originalName,
		savedName,
		storageKey,
		mimeType,
		size,
		now,
	)

	if file.ID() != fileID {
		t.Fatalf("expected ID %q, got %q", fileID, file.ID())
	}

	if file.ChannelID() != channelID {
		t.Fatalf("expected channel ID %q, got %q", channelID, file.ChannelID())
	}

	if file.UserID() != userID {
		t.Fatalf("expected user ID %q, got %q", userID, file.UserID())
	}

	if file.OriginalName() != originalName {
		t.Fatalf("expected original name %q, got %q", originalName, file.OriginalName())
	}

	if file.SavedName() != savedName {
		t.Fatalf("expected saved name %q, got %q", savedName, file.SavedName())
	}

	if file.StorageKey() != storageKey {
		t.Fatalf("expected storage key %q, got %q", storageKey, file.StorageKey())
	}

	if file.MimeType() != mimeType {
		t.Fatalf("expected MIME type %q, got %q", mimeType, file.MimeType())
	}

	if file.Size() != size {
		t.Fatalf("expected size %d, got %d", size, file.Size())
	}

	if !file.UploadedAt().Equal(now) {
		t.Fatalf("expected uploadedAt %v, got %v", now, file.UploadedAt())
	}

	events := file.PullEvents()

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event, ok := events[0].(FileCreatedEvent)
	if !ok {
		t.Fatalf("expected FileCreatedEvent, got %T", events[0])
	}

	if event.FileID != fileID {
		t.Fatalf("expected event file ID %q, got %q", fileID, event.FileID)
	}

	if event.ChannelID != channelID {
		t.Fatalf("expected event channel ID %q, got %q", channelID, event.ChannelID)
	}

	if event.UserID != userID {
		t.Fatalf("expected event user ID %q, got %q", userID, event.UserID)
	}

	if event.OriginalName != originalName {
		t.Fatalf("expected event original name %q, got %q", originalName, event.OriginalName)
	}

	if event.StorageKey != storageKey {
		t.Fatalf("expected event storage key %q, got %q", storageKey, event.StorageKey)
	}

	if event.MimeType != mimeType {
		t.Fatalf("expected event MIME type %q, got %q", mimeType, event.MimeType)
	}

	if event.Size != size {
		t.Fatalf("expected event size %d, got %d", size, event.Size)
	}

	if !event.OccurredAt.Equal(now) {
		t.Fatalf("expected event occurredAt %v, got %v", now, event.OccurredAt)
	}

	if event.EventType() != EventTypeFileCreated {
		t.Fatalf("expected event type %q, got %q", EventTypeFileCreated, event.EventType())
	}
}

func TestRestoreFile(t *testing.T) {
	uploadedAt := time.Now()

	fileID, err := NewFileID("file-1")
	if err != nil {
		t.Fatal(err)
	}

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	userID, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	originalName, err := NewFileName("document.pdf")
	if err != nil {
		t.Fatal(err)
	}

	savedName, err := NewFileName("stored.pdf")
	if err != nil {
		t.Fatal(err)
	}

	file := RestoreFile(
		fileID,
		channelID,
		userID,
		originalName,
		savedName,
		"storage/document.pdf",
		"application/pdf",
		2048,
		uploadedAt,
	)

	if file.ID() != fileID {
		t.Fatalf("expected ID %q, got %q", fileID, file.ID())
	}

	if file.ChannelID() != channelID {
		t.Fatalf("expected channel ID %q, got %q", channelID, file.ChannelID())
	}

	if file.UserID() != userID {
		t.Fatalf("expected user ID %q, got %q", userID, file.UserID())
	}

	if !file.UploadedAt().Equal(uploadedAt) {
		t.Fatalf("expected uploadedAt %v, got %v", uploadedAt, file.UploadedAt())
	}

	if events := file.PullEvents(); len(events) != 0 {
		t.Fatalf("restored file should have no pending events, got %d", len(events))
	}
}

func TestFilePullEventsClearsEvents(t *testing.T) {
	now := time.Now()

	fileID, err := NewFileID("file-1")
	if err != nil {
		t.Fatal(err)
	}

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	userID, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	originalName, err := NewFileName("document.pdf")
	if err != nil {
		t.Fatal(err)
	}

	savedName, err := NewFileName("stored.pdf")
	if err != nil {
		t.Fatal(err)
	}

	file := NewFile(
		fileID,
		channelID,
		userID,
		originalName,
		savedName,
		"storage/document.pdf",
		"application/pdf",
		1024,
		now,
	)

	first := file.PullEvents()

	if len(first) != 1 {
		t.Fatalf("expected 1 event from first pull, got %d", len(first))
	}

	second := file.PullEvents()

	if len(second) != 0 {
		t.Fatalf("expected no events after clearing, got %d", len(second))
	}
}
