package task

import (
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	"time"
)

type Task struct {
	id               TaskID
	projectID        project.ProjectID
	priority         Priority
	status           Status
	name             Name
	description      string
	estimatedMinutes Minutes
	assignees        map[iam.UserID]struct{}

	events []common.Event

	version   int
	createdAt time.Time
	updatedAt time.Time
}

func NewTask(
	id TaskID,
	projectID project.ProjectID,
	priority Priority,
	status Status,
	name Name,
	description string,
	estMinutes Minutes,
	assignees []iam.UserID) (*Task, error) {

	assigneesSet := make(map[iam.UserID]struct{})

	// Adding assignees without duplicates
	for i := range assignees {
		assigneesSet[assignees[i]] = struct{}{}
	}

	now := time.Now()

	return &Task{
		id:               id,
		projectID:        projectID,
		priority:         priority,
		status:           status,
		name:             name,
		description:      description,
		estimatedMinutes: estMinutes,
		assignees:        assigneesSet,
		version:          1,
		createdAt:        now,
		updatedAt:        now,
	}, nil
}

func RestoreTask(
	id TaskID,
	projectID project.ProjectID,
	priority Priority,
	status Status,
	name Name,
	description string,
	estMinutes Minutes,
	assignees []iam.UserID,
	version int,
	createdAt time.Time,
	updatedAt time.Time,
) *Task {

	aMap := make(map[iam.UserID]struct{}, len(assignees))

	return &Task{
		id:               id,
		projectID:        projectID,
		priority:         priority,
		status:           status,
		name:             name,
		description:      description,
		estimatedMinutes: estMinutes,
		assignees:        aMap,
		version:          version,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
	}
}

func (t *Task) ID() TaskID {
	return t.id
}

func (t *Task) ProjectID() project.ProjectID {
	return t.projectID
}

func (t *Task) Priority() Priority {
	return t.priority
}

func (t *Task) Status() Status {
	return t.status
}

func (t *Task) Name() Name {
	return t.name
}

func (t *Task) Description() string {
	return t.description
}

func (t *Task) EstimatedMinutes() Minutes {
	return t.estimatedMinutes
}

func (t *Task) Assignees() []iam.UserID {

	// Pre-allocate the slice to avoid the dynamic growth that occurs
	// when using slices.Collect(maps.Keys(p.members)).
	m := make([]iam.UserID, len(t.assignees))

	i := 0
	for id := range t.assignees {
		m[i] = id
		i++
	}

	return m
}

func (t *Task) ReplaceAssignees(assignees []iam.UserID, now time.Time) {

	m := make(map[iam.UserID]struct{}, len(assignees))
	for i := range assignees {
		m[assignees[i]] = struct{}{}
	}

	t.assignees = m
	t.updatedAt = now

}

func (t *Task) HasAssignee(userID iam.UserID) bool {
	_, ok := t.assignees[userID]
	return ok
}

func (t *Task) Version() int {
	return t.version
}

func (t *Task) CreatedAt() time.Time {
	return t.createdAt
}

func (t *Task) UpdatedAt() time.Time {
	return t.updatedAt
}

func (t *Task) MoveToBacklog(now time.Time) error {
	if t.status == StatusBacklog {
		return ErrTaskAlreadyBacklog
	}

	fromStatus := t.status
	t.status = StatusBacklog
	t.updatedAt = now

	t.events = append(t.events, TaskStatusChangedEvent{
		TaskID:     t.id,
		ProjectID:  t.projectID,
		Name:       t.name,
		FromStatus: fromStatus,
		NewStatus:  t.status,
		OccurredAt: now,
	})

	return nil
}

func (t *Task) MoveToInProgress(now time.Time) error {
	if t.status == StatusInProgress {
		return ErrTaskAlreadyInProgress
	}

	fromStatus := t.status
	t.status = StatusInProgress
	t.updatedAt = now

	t.events = append(t.events, TaskStatusChangedEvent{
		TaskID:     t.id,
		ProjectID:  t.projectID,
		Name:       t.name,
		FromStatus: fromStatus,
		NewStatus:  t.status,
		OccurredAt: now,
	})

	return nil
}

func (t *Task) MoveToCompleted(now time.Time) error {
	if t.status == StatusCompleted {
		return ErrTaskAlreadyCompleted
	}

	fromStatus := t.status
	t.status = StatusCompleted
	t.updatedAt = now

	t.events = append(t.events, TaskStatusChangedEvent{
		TaskID:     t.id,
		ProjectID:  t.projectID,
		Name:       t.name,
		FromStatus: fromStatus,
		NewStatus:  t.status,
		OccurredAt: now,
	})

	return nil
}

func (t *Task) PullEvents() []common.Event {
	events := t.events
	t.events = nil

	return events
}

func (t *Task) Equals(task *Task) bool {
	return t.id == task.ID()
}
