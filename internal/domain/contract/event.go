package contract

import (
	"mizu/internal/domain/common"
	"mizu/internal/domain/project"
	"time"
)

var (
	EventTypeContractCreated       common.EventType = "contract.created"
	EventTypeContractStatusChanged common.EventType = "contract.status.changed"
)

type ContractCreatedEvent struct {
	ContractID ContractID
	ProjectID  project.ProjectID
	Name       Name
	OccurredAt time.Time
}

func (e ContractCreatedEvent) EventType() common.EventType {
	return EventTypeContractCreated
}

type ContractStatusChangedEvent struct {
	ContractID ContractID
	ProjectID  project.ProjectID
	Name       Name
	Signatory  Signatory
	Status     Status
	OccurredAt time.Time
}

func (e ContractStatusChangedEvent) EventType() common.EventType {
	return EventTypeContractStatusChanged
}
