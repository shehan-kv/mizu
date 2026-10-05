package task

import (
	"testing"
	"time"

	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

func newTestTask(t *testing.T, now time.Time, status Status) *Task {
	t.Helper()

	taskID, err := NewTaskID("task-1")
	if err != nil {
		t.Fatalf("NewTaskID() error = %v", err)
	}

	projectID, err := project.NewProjectID("project-1")
	if err != nil {
		t.Fatalf("NewProjectID() error = %v", err)
	}

	priority, err := NewPriority("medium")
	if err != nil {
		t.Fatalf("NewPriority() error = %v", err)
	}

	name, err := NewName("Implement authentication")
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}

	minutes, err := NewMinutes(120)
	if err != nil {
		t.Fatalf("NewMinutes() error = %v", err)
	}

	task, err := NewTask(
		taskID,
		projectID,
		priority,
		status,
		name,
		"Implement user authentication and session handling.",
		minutes,
		[]iam.UserID{"user-1", "user-2"},
		now,
	)
	if err != nil {
		t.Fatalf("NewTask() error = %v", err)
	}

	return task
}

func TestNewTask(t *testing.T) {
	now := time.Now()

	taskID, err := NewTaskID("task-1")
	if err != nil {
		t.Fatalf("NewTaskID() error = %v", err)
	}

	projectID, err := project.NewProjectID("project-1")
	if err != nil {
		t.Fatalf("NewProjectID() error = %v", err)
	}

	priority, err := NewPriority("high")
	if err != nil {
		t.Fatalf("NewPriority() error = %v", err)
	}

	status, err := NewStatus("in-progress")
	if err != nil {
		t.Fatalf("NewStatus() error = %v", err)
	}

	name, err := NewName("Build dashboard")
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}

	minutes, err := NewMinutes(90)
	if err != nil {
		t.Fatalf("NewMinutes() error = %v", err)
	}

	task, err := NewTask(
		taskID,
		projectID,
		priority,
		status,
		name,
		"Build the project dashboard.",
		minutes,
		[]iam.UserID{"user-1", "user-1", "user-2"},
		now,
	)
	if err != nil {
		t.Fatalf("NewTask() error = %v", err)
	}

	if task.ID() != taskID {
		t.Fatalf("ID() = %v, want %v", task.ID(), taskID)
	}

	if task.ProjectID() != projectID {
		t.Fatalf("ProjectID() = %v, want %v", task.ProjectID(), projectID)
	}

	if task.Priority() != priority {
		t.Fatalf("Priority() = %v, want %v", task.Priority(), priority)
	}

	if task.Status() != status {
		t.Fatalf("Status() = %v, want %v", task.Status(), status)
	}

	if task.Name() != name {
		t.Fatalf("Name() = %v, want %v", task.Name(), name)
	}

	if task.Description() != "Build the project dashboard." {
		t.Fatalf("unexpected description: %q", task.Description())
	}

	if task.EstimatedMinutes() != minutes {
		t.Fatalf(
			"EstimatedMinutes() = %v, want %v",
			task.EstimatedMinutes(),
			minutes,
		)
	}

	if len(task.Assignees()) != 2 {
		t.Fatalf("expected 2 unique assignees, got %d", len(task.Assignees()))
	}

	if !task.HasAssignee("user-1") {
		t.Fatal("user-1 should be an assignee")
	}

	if !task.HasAssignee("user-2") {
		t.Fatal("user-2 should be an assignee")
	}

	if task.Version() != 1 {
		t.Fatalf("Version() = %d, want 1", task.Version())
	}

	if !task.CreatedAt().Equal(now) {
		t.Fatalf("CreatedAt() = %v, want %v", task.CreatedAt(), now)
	}

	if !task.UpdatedAt().Equal(now) {
		t.Fatalf("UpdatedAt() = %v, want %v", task.UpdatedAt(), now)
	}

	if !task.CreatedAt().Equal(task.UpdatedAt()) {
		t.Fatal("CreatedAt() and UpdatedAt() should initially be equal")
	}

	if events := task.PullEvents(); len(events) != 0 {
		t.Fatalf("new task should have no events, got %d", len(events))
	}
}

func TestRestoreTask(t *testing.T) {
	createdAt := time.Now()
	updatedAt := createdAt.Add(time.Second)

	taskID, err := NewTaskID("task-1")
	if err != nil {
		t.Fatalf("NewTaskID() error = %v", err)
	}

	projectID, err := project.NewProjectID("project-1")
	if err != nil {
		t.Fatalf("NewProjectID() error = %v", err)
	}

	priority, err := NewPriority("low")
	if err != nil {
		t.Fatalf("NewPriority() error = %v", err)
	}

	status, err := NewStatus("completed")
	if err != nil {
		t.Fatalf("NewStatus() error = %v", err)
	}

	name, err := NewName("Review pull request")
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}

	minutes, err := NewMinutes(30)
	if err != nil {
		t.Fatalf("NewMinutes() error = %v", err)
	}

	task := RestoreTask(
		taskID,
		projectID,
		priority,
		status,
		name,
		"Review the completed implementation.",
		minutes,
		[]iam.UserID{"user-1", "user-2"},
		7,
		createdAt,
		updatedAt,
	)

	if task.ID() != taskID {
		t.Fatalf("ID() = %v, want %v", task.ID(), taskID)
	}

	if task.ProjectID() != projectID {
		t.Fatalf("ProjectID() = %v, want %v", task.ProjectID(), projectID)
	}

	if task.Priority() != priority {
		t.Fatalf("Priority() = %v, want %v", task.Priority(), priority)
	}

	if task.Status() != status {
		t.Fatalf("Status() = %v, want %v", task.Status(), status)
	}

	if task.Name() != name {
		t.Fatalf("Name() = %v, want %v", task.Name(), name)
	}

	if task.Description() != "Review the completed implementation." {
		t.Fatalf("unexpected description: %q", task.Description())
	}

	if task.EstimatedMinutes() != minutes {
		t.Fatalf(
			"EstimatedMinutes() = %v, want %v",
			task.EstimatedMinutes(),
			minutes,
		)
	}

	if task.Version() != 7 {
		t.Fatalf("Version() = %d, want 7", task.Version())
	}

	if !task.CreatedAt().Equal(createdAt) {
		t.Fatalf("CreatedAt() = %v, want %v", task.CreatedAt(), createdAt)
	}

	if !task.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("UpdatedAt() = %v, want %v", task.UpdatedAt(), updatedAt)
	}

	if len(task.Assignees()) != 2 {
		t.Fatalf("expected 2 assignees, got %d", len(task.Assignees()))
	}

	if !task.HasAssignee("user-1") {
		t.Fatal("user-1 should be restored as an assignee")
	}

	if !task.HasAssignee("user-2") {
		t.Fatal("user-2 should be restored as an assignee")
	}

	if events := task.PullEvents(); len(events) != 0 {
		t.Fatalf("restored task should have no events, got %d", len(events))
	}
}

func TestTaskAssignees(t *testing.T) {
	now := time.Now()
	task := newTestTask(t, now, StatusBacklog)

	t.Run("returns current assignees", func(t *testing.T) {
		assignees := task.Assignees()

		if len(assignees) != 2 {
			t.Fatalf("expected 2 assignees, got %d", len(assignees))
		}

		if !task.HasAssignee("user-1") {
			t.Fatal("user-1 should be an assignee")
		}

		if !task.HasAssignee("user-2") {
			t.Fatal("user-2 should be an assignee")
		}
	})

	t.Run("returns a copy", func(t *testing.T) {
		assignees := task.Assignees()
		assignees[0] = "modified"

		if task.HasAssignee("modified") {
			t.Fatal("modifying returned slice changed task state")
		}

		if len(task.Assignees()) != 2 {
			t.Fatal("task assignees were unexpectedly modified")
		}
	})

	t.Run("checks whether user is assigned", func(t *testing.T) {
		if !task.HasAssignee("user-1") {
			t.Fatal("user-1 should be assigned")
		}

		if task.HasAssignee("unknown") {
			t.Fatal("unknown user should not be assigned")
		}
	})
}

func TestTaskReplaceAssignees(t *testing.T) {
	now := time.Now()
	task := newTestTask(t, now, StatusBacklog)

	updatedAt := now.Add(time.Second)

	task.ReplaceAssignees(
		[]iam.UserID{"user-3", "user-3", "user-4"},
		updatedAt,
	)

	if len(task.Assignees()) != 2 {
		t.Fatalf("expected 2 assignees, got %d", len(task.Assignees()))
	}

	if !task.HasAssignee("user-3") {
		t.Fatal("user-3 should be an assignee")
	}

	if !task.HasAssignee("user-4") {
		t.Fatal("user-4 should be an assignee")
	}

	if task.HasAssignee("user-1") {
		t.Fatal("user-1 should no longer be an assignee")
	}

	if task.HasAssignee("user-2") {
		t.Fatal("user-2 should no longer be an assignee")
	}

	if !task.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("UpdatedAt() = %v, want %v", task.UpdatedAt(), updatedAt)
	}
}

func TestTaskMoveToBacklog(t *testing.T) {
	now := time.Now()
	task := newTestTask(t, now, StatusInProgress)

	updatedAt := now.Add(time.Second)

	err := task.MoveToBacklog(updatedAt)
	if err != nil {
		t.Fatalf("MoveToBacklog() error = %v", err)
	}

	if task.Status() != StatusBacklog {
		t.Fatalf("Status() = %v, want %v", task.Status(), StatusBacklog)
	}

	if !task.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("UpdatedAt() = %v, want %v", task.UpdatedAt(), updatedAt)
	}

	assertTaskStatusChangedEvent(
		t,
		task,
		StatusInProgress,
		StatusBacklog,
		updatedAt,
	)
}

func TestTaskMoveToBacklogWhenAlreadyBacklog(t *testing.T) {
	now := time.Now()
	task := newTestTask(t, now, StatusBacklog)

	err := task.MoveToBacklog(now.Add(time.Second))
	if err != ErrTaskAlreadyBacklog {
		t.Fatalf("expected ErrTaskAlreadyBacklog, got %v", err)
	}

	if task.Status() != StatusBacklog {
		t.Fatalf("Status() = %v, want %v", task.Status(), StatusBacklog)
	}

	if !task.UpdatedAt().Equal(now) {
		t.Fatal("UpdatedAt() changed after rejected transition")
	}

	if events := task.PullEvents(); len(events) != 0 {
		t.Fatalf("expected no events, got %d", len(events))
	}
}

func TestTaskMoveToInProgress(t *testing.T) {
	now := time.Now()
	task := newTestTask(t, now, StatusBacklog)

	updatedAt := now.Add(time.Second)

	err := task.MoveToInProgress(updatedAt)
	if err != nil {
		t.Fatalf("MoveToInProgress() error = %v", err)
	}

	if task.Status() != StatusInProgress {
		t.Fatalf("Status() = %v, want %v", task.Status(), StatusInProgress)
	}

	if !task.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("UpdatedAt() = %v, want %v", task.UpdatedAt(), updatedAt)
	}

	assertTaskStatusChangedEvent(
		t,
		task,
		StatusBacklog,
		StatusInProgress,
		updatedAt,
	)
}

func TestTaskMoveToInProgressWhenAlreadyInProgress(t *testing.T) {
	now := time.Now()
	task := newTestTask(t, now, StatusInProgress)

	updatedAt := now.Add(time.Second)

	err := task.MoveToInProgress(updatedAt)
	if err != ErrTaskAlreadyInProgress {
		t.Fatalf("expected ErrTaskAlreadyInProgress, got %v", err)
	}

	if task.Status() != StatusInProgress {
		t.Fatalf("Status() = %v, want %v", task.Status(), StatusInProgress)
	}

	if !task.UpdatedAt().Equal(now) {
		t.Fatal("UpdatedAt() changed after rejected transition")
	}

	if events := task.PullEvents(); len(events) != 0 {
		t.Fatalf("expected no events, got %d", len(events))
	}
}

func TestTaskMoveToCompleted(t *testing.T) {
	now := time.Now()
	task := newTestTask(t, now, StatusBacklog)

	updatedAt := now.Add(time.Second)

	err := task.MoveToCompleted(updatedAt)
	if err != nil {
		t.Fatalf("MoveToCompleted() error = %v", err)
	}

	if task.Status() != StatusCompleted {
		t.Fatalf("Status() = %v, want %v", task.Status(), StatusCompleted)
	}

	if !task.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("UpdatedAt() = %v, want %v", task.UpdatedAt(), updatedAt)
	}

	assertTaskStatusChangedEvent(
		t,
		task,
		StatusBacklog,
		StatusCompleted,
		updatedAt,
	)
}

func TestTaskMoveToCompletedWhenAlreadyCompleted(t *testing.T) {
	now := time.Now()
	task := newTestTask(t, now, StatusCompleted)

	updatedAt := now.Add(time.Second)

	err := task.MoveToCompleted(updatedAt)
	if err != ErrTaskAlreadyCompleted {
		t.Fatalf("expected ErrTaskAlreadyCompleted, got %v", err)
	}

	if task.Status() != StatusCompleted {
		t.Fatalf("Status() = %v, want %v", task.Status(), StatusCompleted)
	}

	if !task.UpdatedAt().Equal(now) {
		t.Fatal("UpdatedAt() changed after rejected transition")
	}

	if events := task.PullEvents(); len(events) != 0 {
		t.Fatalf("expected no events, got %d", len(events))
	}
}

func TestTaskStatusTransitionsCanMoveFromAnyOtherStatus(t *testing.T) {
	now := time.Now()
	task := newTestTask(t, now, StatusCompleted)

	if err := task.MoveToBacklog(now.Add(time.Second)); err != nil {
		t.Fatalf("MoveToBacklog() error = %v", err)
	}

	if task.Status() != StatusBacklog {
		t.Fatalf("Status() = %v, want %v", task.Status(), StatusBacklog)
	}

	if events := task.PullEvents(); len(events) != 1 {
		t.Fatalf("expected 1 status event, got %d", len(events))
	}
}

func TestTaskPullEvents(t *testing.T) {
	now := time.Now()
	task := newTestTask(t, now, StatusBacklog)

	firstAt := now.Add(time.Second)
	secondAt := now.Add(2 * time.Second)

	if err := task.MoveToInProgress(firstAt); err != nil {
		t.Fatalf("MoveToInProgress() error = %v", err)
	}

	if err := task.MoveToCompleted(secondAt); err != nil {
		t.Fatalf("MoveToCompleted() error = %v", err)
	}

	events := task.PullEvents()

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	first, ok := events[0].(TaskStatusChangedEvent)
	if !ok {
		t.Fatalf("expected TaskStatusChangedEvent, got %T", events[0])
	}

	if first.TaskID != task.ID() {
		t.Fatalf("first event TaskID = %v, want %v", first.TaskID, task.ID())
	}

	if first.ProjectID != task.ProjectID() {
		t.Fatalf(
			"first event ProjectID = %v, want %v",
			first.ProjectID,
			task.ProjectID(),
		)
	}

	if first.Name != task.Name() {
		t.Fatalf("first event Name = %v, want %v", first.Name, task.Name())
	}

	if first.FromStatus != StatusBacklog {
		t.Fatalf(
			"first event FromStatus = %v, want %v",
			first.FromStatus,
			StatusBacklog,
		)
	}

	if first.NewStatus != StatusInProgress {
		t.Fatalf(
			"first event NewStatus = %v, want %v",
			first.NewStatus,
			StatusInProgress,
		)
	}

	if !first.OccurredAt.Equal(firstAt) {
		t.Fatalf(
			"first event OccurredAt = %v, want %v",
			first.OccurredAt,
			firstAt,
		)
	}

	second, ok := events[1].(TaskStatusChangedEvent)
	if !ok {
		t.Fatalf("expected TaskStatusChangedEvent, got %T", events[1])
	}

	if second.TaskID != task.ID() {
		t.Fatalf("second event TaskID = %v, want %v", second.TaskID, task.ID())
	}

	if second.ProjectID != task.ProjectID() {
		t.Fatalf(
			"second event ProjectID = %v, want %v",
			second.ProjectID,
			task.ProjectID(),
		)
	}

	if second.Name != task.Name() {
		t.Fatalf("second event Name = %v, want %v", second.Name, task.Name())
	}

	if second.FromStatus != StatusInProgress {
		t.Fatalf(
			"second event FromStatus = %v, want %v",
			second.FromStatus,
			StatusInProgress,
		)
	}

	if second.NewStatus != StatusCompleted {
		t.Fatalf(
			"second event NewStatus = %v, want %v",
			second.NewStatus,
			StatusCompleted,
		)
	}

	if !second.OccurredAt.Equal(secondAt) {
		t.Fatalf(
			"second event OccurredAt = %v, want %v",
			second.OccurredAt,
			secondAt,
		)
	}

	if remaining := task.PullEvents(); len(remaining) != 0 {
		t.Fatalf("expected events to be cleared, got %d", len(remaining))
	}
}

func TestTaskEquals(t *testing.T) {
	now := time.Now()

	first := newTestTask(t, now, StatusBacklog)
	sameID := newTestTask(t, now.Add(time.Second), StatusCompleted)

	differentID := RestoreTask(
		"task-2",
		first.ProjectID(),
		first.Priority(),
		first.Status(),
		first.Name(),
		first.Description(),
		first.EstimatedMinutes(),
		first.Assignees(),
		first.Version(),
		first.CreatedAt(),
		first.UpdatedAt(),
	)

	if !first.Equals(sameID) {
		t.Fatal("tasks with the same ID should be equal")
	}

	if first.Equals(differentID) {
		t.Fatal("tasks with different IDs should not be equal")
	}
}

func assertTaskStatusChangedEvent(
	t *testing.T,
	task *Task,
	wantFrom Status,
	wantNew Status,
	wantOccurredAt time.Time,
) {
	t.Helper()

	events := task.PullEvents()

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event, ok := events[0].(TaskStatusChangedEvent)
	if !ok {
		t.Fatalf("expected TaskStatusChangedEvent, got %T", events[0])
	}

	if event.EventType() != EventTypeTaskStatusChanged {
		t.Fatalf(
			"EventType() = %v, want %v",
			event.EventType(),
			EventTypeTaskStatusChanged,
		)
	}

	if event.TaskID != task.ID() {
		t.Fatalf("TaskID = %v, want %v", event.TaskID, task.ID())
	}

	if event.ProjectID != task.ProjectID() {
		t.Fatalf("ProjectID = %v, want %v", event.ProjectID, task.ProjectID())
	}

	if event.Name != task.Name() {
		t.Fatalf("Name = %v, want %v", event.Name, task.Name())
	}

	if event.FromStatus != wantFrom {
		t.Fatalf("FromStatus = %v, want %v", event.FromStatus, wantFrom)
	}

	if event.NewStatus != wantNew {
		t.Fatalf("NewStatus = %v, want %v", event.NewStatus, wantNew)
	}

	if !event.OccurredAt.Equal(wantOccurredAt) {
		t.Fatalf(
			"OccurredAt = %v, want %v",
			event.OccurredAt,
			wantOccurredAt,
		)
	}
}
