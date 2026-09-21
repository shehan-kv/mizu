package contract

import (
	"testing"
	"time"

	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

func TestContractCreatedEventEventType(t *testing.T) {
	event := ContractCreatedEvent{
		ContractID: ContractID("contract-1"),
		ProjectID:  project.ProjectID("project-1"),
		Name:       Name{value: "Test Contract"},
		OccurredAt: time.Now(),
	}

	if event.EventType() != EventTypeContractCreated {
		t.Fatalf(
			"expected %q, got %q",
			EventTypeContractCreated,
			event.EventType(),
		)
	}
}

func TestContractStatusChangedEventEventType(t *testing.T) {
	now := time.Now()

	event := ContractStatusChangedEvent{
		ContractID: ContractID("contract-1"),
		ProjectID:  project.ProjectID("project-1"),
		Name:       Name{value: "Test Contract"},
		Signatory:  NewSignatory(iam.UserID("user-1"), now),
		Status:     StatusSigned,
		OccurredAt: now,
	}

	if event.EventType() != EventTypeContractStatusChanged {
		t.Fatalf(
			"expected %q, got %q",
			EventTypeContractStatusChanged,
			event.EventType(),
		)
	}
}
