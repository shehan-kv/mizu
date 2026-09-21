package iam

import (
	"testing"
	"time"
)

func TestUserCreatedEventEventType(t *testing.T) {
	event := UserCreatedEvent{}

	if got := event.EventType(); got != EventTypeUserCreated {
		t.Errorf("expected %q, got %q", EventTypeUserCreated, got)
	}
}

func TestUserVerifiedEventEventType(t *testing.T) {
	event := UserVerifiedEvent{}

	if got := event.EventType(); got != EventTypeUserVerified {
		t.Errorf("expected %q, got %q", EventTypeUserVerified, got)
	}
}

func TestUserEmailChangedEventEventType(t *testing.T) {
	event := UserEmailChangedEvent{}

	if got := event.EventType(); got != EventTypeUserEmailChanged {
		t.Errorf("expected %q, got %q", EventTypeUserEmailChanged, got)
	}
}

func TestRecoveryCreatedEventEventType(t *testing.T) {
	event := RecoveryCreatedEvent{}

	if got := event.EventType(); got != EventTypeRecoveryCreated {
		t.Errorf("expected %q, got %q", EventTypeRecoveryCreated, got)
	}
}

func TestEventTimestampsCanBeCompared(t *testing.T) {
	now := time.Now()

	event := UserVerifiedEvent{
		OccurredAt: now,
	}

	if !event.OccurredAt.Equal(now) {
		t.Errorf("expected event timestamp %v, got %v", now, event.OccurredAt)
	}
}
