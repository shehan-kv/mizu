package iam

import (
	"testing"
	"time"
)

func TestNewRecovery(t *testing.T) {
	userID := UserID("user-123")
	token := RecoveryToken("recovery-token")
	now := time.Now()

	recovery := NewRecovery(userID, token, now)

	if recovery.UserID() != userID {
		t.Errorf("expected user ID %q, got %q", userID, recovery.UserID())
	}

	if recovery.Token() != token {
		t.Errorf("expected token %q, got %q", token, recovery.Token())
	}

	if recovery.Version() != 1 {
		t.Errorf("expected version 1, got %d", recovery.Version())
	}

	if !recovery.CreatedAt().Equal(now) {
		t.Errorf("expected createdAt %v, got %v", now, recovery.CreatedAt())
	}
}

func TestNewRecoveryCreatesRecoveryCreatedEvent(t *testing.T) {
	userID := UserID("user-123")
	token := RecoveryToken("recovery-token")
	now := time.Now()

	recovery := NewRecovery(userID, token, now)

	events := recovery.PullEvents()

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event, ok := events[0].(RecoveryCreatedEvent)
	if !ok {
		t.Fatalf("expected RecoveryCreatedEvent, got %T", events[0])
	}

	if event.UserID != userID {
		t.Errorf("expected user ID %q, got %q", userID, event.UserID)
	}

	if event.Token != token {
		t.Errorf("expected token %q, got %q", token, event.Token)
	}

	if !event.OccurredAt.Equal(now) {
		t.Errorf("expected occurredAt %v, got %v", now, event.OccurredAt)
	}
}

func TestRecoveryPullEventsClearsEvents(t *testing.T) {
	now := time.Now()

	recovery := NewRecovery(
		UserID("user-123"),
		RecoveryToken("token"),
		now,
	)

	firstPull := recovery.PullEvents()

	if len(firstPull) != 1 {
		t.Fatalf("expected first pull to contain 1 event, got %d", len(firstPull))
	}

	secondPull := recovery.PullEvents()

	if len(secondPull) != 0 {
		t.Fatalf("expected second pull to contain 0 events, got %d", len(secondPull))
	}
}

func TestRestoreRecovery(t *testing.T) {
	userID := UserID("user-123")
	token := RecoveryToken("token")
	version := 4
	createdAt := time.Now()

	recovery := RestoreRecovery(
		userID,
		token,
		version,
		createdAt,
	)

	if recovery.UserID() != userID {
		t.Errorf("expected user ID %q, got %q", userID, recovery.UserID())
	}

	if recovery.Token() != token {
		t.Errorf("expected token %q, got %q", token, recovery.Token())
	}

	if recovery.Version() != version {
		t.Errorf("expected version %d, got %d", version, recovery.Version())
	}

	if !recovery.CreatedAt().Equal(createdAt) {
		t.Errorf("expected createdAt %v, got %v", createdAt, recovery.CreatedAt())
	}

	if events := recovery.PullEvents(); len(events) != 0 {
		t.Errorf("expected restored recovery to have no events, got %d", len(events))
	}
}
