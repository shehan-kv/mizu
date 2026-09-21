package project

import (
	"errors"
	"testing"
	"time"

	"mizu/internal/domain/iam"
)

func newTestProject(t *testing.T, now time.Time) *Project {
	t.Helper()

	id, err := NewProjectID("project-1")
	if err != nil {
		t.Fatalf("failed to create project ID: %v", err)
	}

	name, err := NewName("Test Project")
	if err != nil {
		t.Fatalf("failed to create project name: %v", err)
	}

	status, err := NewStatus("started")
	if err != nil {
		t.Fatalf("failed to create project status: %v", err)
	}

	createdBy := iam.UserID("admin-1")

	project, err := NewProject(
		id,
		name,
		status,
		createdBy,
		[]iam.UserID{
			iam.UserID("user-1"),
			iam.UserID("user-2"),
		},
		now,
	)
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	return project
}

func TestNewProject(t *testing.T) {
	now := time.Now()

	t.Run("creates project", func(t *testing.T) {
		project := newTestProject(t, now)

		if project.ID() != ProjectID("project-1") {
			t.Fatalf("unexpected project ID: %q", project.ID())
		}

		if project.Name().String() != "Test Project" {
			t.Fatalf("unexpected project name: %q", project.Name())
		}

		if project.Status() != StatusStarted {
			t.Fatalf(
				"expected status %q, got %q",
				StatusStarted,
				project.Status(),
			)
		}

		if project.Version() != 1 {
			t.Fatalf("expected version 1, got %d", project.Version())
		}

		if !project.CreatedAt().Equal(now) {
			t.Fatalf(
				"expected CreatedAt %v, got %v",
				now,
				project.CreatedAt(),
			)
		}

		if !project.UpdatedAt().Equal(now) {
			t.Fatalf(
				"expected UpdatedAt %v, got %v",
				now,
				project.UpdatedAt(),
			)
		}
	})

	t.Run("creator is automatically added as member", func(t *testing.T) {
		project := newTestProject(t, now)

		if !project.HasMember(iam.UserID("admin-1")) {
			t.Fatal("expected creator to be a project member")
		}
	})

	t.Run("adds supplied members", func(t *testing.T) {
		project := newTestProject(t, now)

		for _, id := range []iam.UserID{
			"user-1",
			"user-2",
		} {
			if !project.HasMember(id) {
				t.Fatalf("expected %q to be a project member", id)
			}
		}
	})

	t.Run("removes duplicate members automatically", func(t *testing.T) {
		id, err := NewProjectID("project-1")
		if err != nil {
			t.Fatalf("failed to create ID: %v", err)
		}

		name, err := NewName("Test Project")
		if err != nil {
			t.Fatalf("failed to create name: %v", err)
		}

		status := StatusStarted
		createdBy := iam.UserID("user-1")

		project, err := NewProject(
			id,
			name,
			status,
			createdBy,
			[]iam.UserID{
				"user-1",
				"user-1",
				"user-2",
				"user-2",
			},
			now,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(project.Members()) != 2 {
			t.Fatalf(
				"expected 2 unique members, got %d",
				len(project.Members()),
			)
		}
	})

	t.Run("rejects empty supplied member ID", func(t *testing.T) {
		id, err := NewProjectID("project-1")
		if err != nil {
			t.Fatalf("failed to create ID: %v", err)
		}

		name, err := NewName("Test Project")
		if err != nil {
			t.Fatalf("failed to create name: %v", err)
		}

		_, err = NewProject(
			id,
			name,
			StatusStarted,
			iam.UserID("user-1"),
			[]iam.UserID{
				"user-2",
				"",
			},
			now,
		)

		if err != ErrProjectMemberIDCannotBeEmpty {
			t.Fatalf(
				"expected %v, got %v",
				ErrProjectMemberIDCannotBeEmpty,
				err,
			)
		}
	})
}

func TestRestoreProject(t *testing.T) {
	createdAt := time.Now()
	updatedAt := createdAt.Add(time.Minute)

	project := RestoreProject(
		ProjectID("project-1"),
		Name{value: "Restored Project"},
		StatusPaused,
		[]iam.UserID{
			"user-1",
			"user-2",
		},
		5,
		createdAt,
		updatedAt,
	)

	if project.ID() != ProjectID("project-1") {
		t.Fatalf("unexpected ID: %q", project.ID())
	}

	if project.Name().String() != "Restored Project" {
		t.Fatalf("unexpected name: %q", project.Name())
	}

	if project.Status() != StatusPaused {
		t.Fatalf("unexpected status: %q", project.Status())
	}

	if project.Version() != 5 {
		t.Fatalf("expected version 5, got %d", project.Version())
	}

	if !project.CreatedAt().Equal(createdAt) {
		t.Fatalf("unexpected CreatedAt")
	}

	if !project.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("unexpected UpdatedAt")
	}

	if !project.HasMember("user-1") ||
		!project.HasMember("user-2") {
		t.Fatal("expected restored members to exist")
	}
}

func TestProjectAcceptsTasks(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{
			name:     "started project accepts tasks",
			status:   StatusStarted,
			expected: true,
		},
		{
			name:     "paused project accepts tasks",
			status:   StatusPaused,
			expected: true,
		},
		{
			name:     "cancelled project does not accept tasks",
			status:   StatusCancelled,
			expected: false,
		},
		{
			name:     "completed project does not accept tasks",
			status:   StatusCompleted,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := RestoreProject(
				ProjectID("project-1"),
				Name{value: "Test Project"},
				tt.status,
				[]iam.UserID{"user-1"},
				1,
				now,
				now,
			)

			if project.AcceptsTasks() != tt.expected {
				t.Fatalf(
					"expected AcceptsTasks() = %v, got %v",
					tt.expected,
					project.AcceptsTasks(),
				)
			}
		})
	}
}

func TestProjectStatusChanges(t *testing.T) {
	t.Run("marks project as started", func(t *testing.T) {
		now := time.Now()
		project := newTestProject(t, now)

		updatedAt := now.Add(time.Second)

		project.MarkAsStarted(updatedAt)

		if project.Status() != StatusStarted {
			t.Fatalf("expected started status")
		}

		if !project.UpdatedAt().Equal(updatedAt) {
			t.Fatalf("expected UpdatedAt to be updated")
		}
	})

	t.Run("marks project as paused", func(t *testing.T) {
		now := time.Now()
		project := newTestProject(t, now)

		updatedAt := now.Add(time.Second)

		project.MarkAsPaused(updatedAt)

		if project.Status() != StatusPaused {
			t.Fatalf("expected paused status")
		}

		if !project.UpdatedAt().Equal(updatedAt) {
			t.Fatalf("expected UpdatedAt to be updated")
		}
	})

	t.Run("marks project as cancelled", func(t *testing.T) {
		now := time.Now()
		project := newTestProject(t, now)

		updatedAt := now.Add(time.Second)

		project.MarkAsCancelled(updatedAt)

		if project.Status() != StatusCancelled {
			t.Fatalf("expected cancelled status")
		}

		if !project.UpdatedAt().Equal(updatedAt) {
			t.Fatalf("expected UpdatedAt to be updated")
		}
	})

	t.Run("marks project as completed", func(t *testing.T) {
		now := time.Now()
		project := newTestProject(t, now)

		updatedAt := now.Add(time.Second)

		project.MarkAsCompleted(updatedAt)

		if project.Status() != StatusCompleted {
			t.Fatalf("expected completed status")
		}

		if !project.UpdatedAt().Equal(updatedAt) {
			t.Fatalf("expected UpdatedAt to be updated")
		}
	})
}

func TestProjectMembers(t *testing.T) {
	now := time.Now()
	project := newTestProject(t, now)

	t.Run("returns current members", func(t *testing.T) {
		members := project.Members()

		if len(members) != 3 {
			t.Fatalf("expected 3 members, got %d", len(members))
		}

		for _, id := range []iam.UserID{
			"admin-1",
			"user-1",
			"user-2",
		} {
			if !project.HasMember(id) {
				t.Fatalf("expected member %q", id)
			}
		}
	})

	t.Run("returns copy of members", func(t *testing.T) {
		members := project.Members()

		members[0] = iam.UserID("modified")

		if project.HasMember("modified") {
			t.Fatal("modifying returned members should not modify project")
		}
	})

	t.Run("checks member existence", func(t *testing.T) {
		if !project.HasMember("user-1") {
			t.Fatal("expected user-1 to be a member")
		}

		if project.HasMember("unknown") {
			t.Fatal("expected unknown user not to be a member")
		}
	})
}

func TestProjectReplaceMembers(t *testing.T) {
	t.Run("replaces members", func(t *testing.T) {
		now := time.Now()
		project := newTestProject(t, now)

		updatedAt := now.Add(time.Second)

		err := project.ReplaceMembers(
			[]iam.UserID{
				"user-3",
				"user-4",
			},
			updatedAt,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(project.Members()) != 2 {
			t.Fatalf("expected 2 members, got %d", len(project.Members()))
		}

		if !project.HasMember("user-3") ||
			!project.HasMember("user-4") {
			t.Fatal("expected replacement members to exist")
		}

		if project.HasMember("user-1") {
			t.Fatal("old member should have been removed")
		}

		if !project.UpdatedAt().Equal(updatedAt) {
			t.Fatal("expected UpdatedAt to be updated")
		}
	})

	t.Run("rejects empty member list", func(t *testing.T) {
		now := time.Now()
		project := newTestProject(t, now)

		err := project.ReplaceMembers(
			[]iam.UserID{},
			now.Add(time.Second),
		)

		if err != ErrProjectMustHaveAtLeastOneMember {
			t.Fatalf(
				"expected %v, got %v",
				ErrProjectMustHaveAtLeastOneMember,
				err,
			)
		}
	})

	t.Run("rejects empty member ID", func(t *testing.T) {
		now := time.Now()
		project := newTestProject(t, now)

		err := project.ReplaceMembers(
			[]iam.UserID{
				"user-3",
				"",
			},
			now.Add(time.Second),
		)

		if err != ErrProjectInvalidMember {
			t.Fatalf(
				"expected %v, got %v",
				ErrProjectInvalidMember,
				err,
			)
		}
	})

	t.Run("deduplicates members", func(t *testing.T) {
		now := time.Now()
		project := newTestProject(t, now)

		err := project.ReplaceMembers(
			[]iam.UserID{
				"user-3",
				"user-3",
				"user-4",
			},
			now.Add(time.Second),
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(project.Members()) != 2 {
			t.Fatalf(
				"expected 2 unique members, got %d",
				len(project.Members()),
			)
		}
	})
}

func TestProjectAddMember(t *testing.T) {
	t.Run("adds member", func(t *testing.T) {
		now := time.Now()
		project := newTestProject(t, now)

		updatedAt := now.Add(time.Second)

		err := project.AddMember("user-3", updatedAt)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !project.HasMember("user-3") {
			t.Fatal("expected user-3 to be added")
		}

		if !project.UpdatedAt().Equal(updatedAt) {
			t.Fatal("expected UpdatedAt to be updated")
		}
	})

	t.Run("rejects empty member ID", func(t *testing.T) {
		now := time.Now()
		project := newTestProject(t, now)

		err := project.AddMember("", now.Add(time.Second))

		if err != ErrProjectMemberIDCannotBeEmpty {
			t.Fatalf(
				"expected %v, got %v",
				ErrProjectMemberIDCannotBeEmpty,
				err,
			)
		}
	})

	t.Run("does not duplicate existing member", func(t *testing.T) {
		now := time.Now()
		project := newTestProject(t, now)

		err := project.AddMember("user-1", now.Add(time.Second))

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(project.Members()) != 3 {
			t.Fatalf(
				"expected 3 members, got %d",
				len(project.Members()),
			)
		}
	})
}

func TestProjectRemoveMember(t *testing.T) {
	now := time.Now()
	t.Run("removes existing member", func(t *testing.T) {
		project := newTestProject(t, now)
		err := project.RemoveMember(iam.UserID("user-1"), now.Add(time.Second))

		if err != nil {
			t.Fatalf("RemoveMember() error = %v", err)
		}

		if project.HasMember(iam.UserID("user-1")) {
			t.Fatal("removed member still exists")
		}

		if !project.HasMember(iam.UserID("user-2")) {
			t.Fatal("existing member was unexpectedly removed")
		}

		if !project.HasMember(iam.UserID("admin-1")) {
			t.Fatal("creator was unexpectedly removed")
		}
	})

	t.Run("removing unknown member is harmless", func(t *testing.T) {
		project := newTestProject(t, now)
		err := project.RemoveMember(iam.UserID("unknown"), now.Add(time.Second))

		if err != nil {
			t.Fatalf("RemoveMember() error = %v", err)
		}

		if len(project.Members()) != 3 {
			t.Fatalf("expected 3 members, got %d", len(project.Members()))
		}
	})

	t.Run("cannot remove last member", func(t *testing.T) {
		project := newTestProject(t, now)
		err := project.ReplaceMembers([]iam.UserID{"only-member"}, now.Add(time.Second))

		if err != nil {
			t.Fatalf("ReplaceMembers() error = %v", err)
		}

		err = project.RemoveMember(iam.UserID("only-member"), now.Add(2*time.Second))

		if !errors.Is(err, ErrProjectMustHaveAtLeastOneMember) {
			t.Fatalf("expected ErrProjectMustHaveAtLeastOneMember, got %v", err)
		}

		if !project.HasMember(iam.UserID("only-member")) {
			t.Fatal("last member was removed despite the error")
		}

		if len(project.Members()) != 1 {
			t.Fatalf("expected 1 member, got %d", len(project.Members()))
		}
	})
}

func TestProjectPullEvents(t *testing.T) {
	now := time.Now()
	project := newTestProject(t, now)

	events := project.PullEvents()

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event, ok := events[0].(ProjectCreatedEvent)
	if !ok {
		t.Fatalf(
			"expected ProjectCreatedEvent, got %T",
			events[0],
		)
	}

	if event.ProjectID != project.ID() {
		t.Fatalf(
			"expected project ID %q, got %q",
			project.ID(),
			event.ProjectID,
		)
	}

	if event.Name != project.Name() {
		t.Fatalf("expected event name to match project name")
	}

	if !event.OccurredAt.Equal(now) {
		t.Fatalf(
			"expected OccurredAt %v, got %v",
			now,
			event.OccurredAt,
		)
	}

	if len(event.Members) != 3 {
		t.Fatalf(
			"expected 3 event members, got %d",
			len(event.Members),
		)
	}

	if event.EventType() != EventTypeProjectCreated {
		t.Fatalf(
			"expected event type %q, got %q",
			EventTypeProjectCreated,
			event.EventType(),
		)
	}

	events = project.PullEvents()

	if len(events) != 0 {
		t.Fatalf(
			"expected events to be cleared, got %d",
			len(events),
		)
	}
}

func TestProjectEquals(t *testing.T) {
	now := time.Now()

	first := newTestProject(t, now)

	second := RestoreProject(
		first.ID(),
		first.Name(),
		first.Status(),
		first.Members(),
		10,
		first.CreatedAt(),
		first.UpdatedAt(),
	)

	if !first.Equals(*second) {
		t.Fatal("projects with the same ID should be equal")
	}

	third := RestoreProject(
		ProjectID("project-2"),
		first.Name(),
		first.Status(),
		first.Members(),
		first.Version(),
		first.CreatedAt(),
		first.UpdatedAt(),
	)

	if first.Equals(*third) {
		t.Fatal("projects with different IDs should not be equal")
	}
}
