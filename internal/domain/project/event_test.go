package project

import (
	"testing"
	"time"

	"mizu/internal/domain/iam"
)

func TestProjectCreatedEventEventType(t *testing.T) {
	event := ProjectCreatedEvent{
		ProjectID: ProjectID("project-1"),
		Name:      Name{value: "Test Project"},
		Members: []iam.UserID{
			iam.UserID("user-1"),
			iam.UserID("user-2"),
		},
		OccurredAt: time.Now(),
	}

	if event.EventType() != EventTypeProjectCreated {
		t.Fatalf(
			"expected %q, got %q",
			EventTypeProjectCreated,
			event.EventType(),
		)
	}
}
