package iam

import (
	"testing"
	"time"
)

func newTestVerificationID(t *testing.T) VerificationID {
	t.Helper()

	id, err := NewVerificationID("verification-1")
	if err != nil {
		t.Fatalf("NewVerificationID() error = %v", err)
	}

	return id
}

func TestNewVerification(t *testing.T) {
	now := time.Now()
	verificationID := newTestVerificationID(t)
	userID := UserID("user-1")

	verification := NewVerification(
		verificationID,
		userID,
		now,
	)

	if verification.ID() != verificationID {
		t.Fatalf(
			"ID() = %v, want %v",
			verification.ID(),
			verificationID,
		)
	}

	if verification.UserID() != userID {
		t.Fatalf(
			"UserID() = %v, want %v",
			verification.UserID(),
			userID,
		)
	}

	if verification.Version() != 1 {
		t.Fatalf(
			"Version() = %d, want 1",
			verification.Version(),
		)
	}

	if !verification.CreatedAt().Equal(now) {
		t.Fatalf(
			"CreatedAt() = %v, want %v",
			verification.CreatedAt(),
			now,
		)
	}

	events := verification.PullEvents()

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event, ok := events[0].(VerificationCreatedEvent)
	if !ok {
		t.Fatalf(
			"expected VerificationCreatedEvent, got %T",
			events[0],
		)
	}

	if event.EventType() != EventTypeVerificationCreated {
		t.Fatalf(
			"EventType() = %v, want %v",
			event.EventType(),
			EventTypeVerificationCreated,
		)
	}

	if event.UserID != userID {
		t.Fatalf(
			"event UserID = %v, want %v",
			event.UserID,
			userID,
		)
	}

	if event.VerificationID != verificationID {
		t.Fatalf(
			"event VerificationID = %v, want %v",
			event.VerificationID,
			verificationID,
		)
	}

	if !event.OccurredAt.Equal(now) {
		t.Fatalf(
			"event OccurredAt = %v, want %v",
			event.OccurredAt,
			now,
		)
	}
}

func TestRestoreVerification(t *testing.T) {
	createdAt := time.Now()
	verificationID := newTestVerificationID(t)
	userID := UserID("user-1")

	verification := RestoreVerification(
		verificationID,
		userID,
		7,
		createdAt,
	)

	if verification.ID() != verificationID {
		t.Fatalf(
			"ID() = %v, want %v",
			verification.ID(),
			verificationID,
		)
	}

	if verification.UserID() != userID {
		t.Fatalf(
			"UserID() = %v, want %v",
			verification.UserID(),
			userID,
		)
	}

	if verification.Version() != 7 {
		t.Fatalf(
			"Version() = %d, want 7",
			verification.Version(),
		)
	}

	if !verification.CreatedAt().Equal(createdAt) {
		t.Fatalf(
			"CreatedAt() = %v, want %v",
			verification.CreatedAt(),
			createdAt,
		)
	}

	if events := verification.PullEvents(); len(events) != 0 {
		t.Fatalf(
			"restored verification should have no events, got %d",
			len(events),
		)
	}
}

func TestVerificationPullEvents(t *testing.T) {
	now := time.Now()
	verification := NewVerification(
		newTestVerificationID(t),
		UserID("user-1"),
		now,
	)

	first := verification.PullEvents()

	if len(first) != 1 {
		t.Fatalf("expected 1 event, got %d", len(first))
	}

	second := verification.PullEvents()

	if len(second) != 0 {
		t.Fatalf("expected events to be cleared, got %d", len(second))
	}
}

func TestVerificationEquals(t *testing.T) {
	now := time.Now()

	firstID := newTestVerificationID(t)
	first := NewVerification(
		firstID,
		UserID("user-1"),
		now,
	)

	sameID := NewVerification(
		firstID,
		UserID("user-2"),
		now.Add(time.Second),
	)

	differentID, err := NewVerificationID("verification-2")
	if err != nil {
		t.Fatalf("NewVerificationID() error = %v", err)
	}

	different := NewVerification(
		differentID,
		UserID("user-1"),
		now,
	)

	if !first.Equals(sameID) {
		t.Fatal("verifications with the same ID should be equal")
	}

	if first.Equals(different) {
		t.Fatal("verifications with different IDs should not be equal")
	}
}
