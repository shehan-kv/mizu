package project

import (
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"time"
)

var (
	EventTypeProjectCreated common.EventType = "project.created"
)

type ProjectCreatedEvent struct {
	ProjectID  ProjectID
	Name       Name
	Members    []iam.UserID
	OccurredAt time.Time
}

func (e ProjectCreatedEvent) EventType() common.EventType {
	return EventTypeProjectCreated
}
func (e ProjectCreatedEvent) EventScope() common.EventScope {
	return common.EventScopeInternal
}
