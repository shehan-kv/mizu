package task

import (
	"mizu/internal/domain/common"
	"mizu/internal/domain/project"
	"time"
)

var (
	EventTypeTaskStatusChanged common.EventType = "task.status.changed"
)

type TaskStatusChangedEvent struct {
	TaskID     TaskID
	ProjectID  project.ProjectID
	Name       Name
	FromStatus Status
	NewStatus  Status
	OccurredAt time.Time
}

func (e TaskStatusChangedEvent) EventType() common.EventType {
	return EventTypeTaskStatusChanged
}
