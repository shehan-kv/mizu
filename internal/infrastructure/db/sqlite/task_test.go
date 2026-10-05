package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	"mizu/internal/domain/task"
)

func TestTaskRepository(t *testing.T) {
	db := newTestSQLite(t)
	repo := NewTaskRepository(db)
	ctx := context.Background()

	t.Run("Add", func(t *testing.T) {
		t.Run("adds task with assignees", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000001",
			)
			userID := newTestTaskUser(
				t,
				db,
				"10000000-0000-0000-0000-000000000002",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			testTask := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000003",
				projectID,
				[]iam.UserID{userID},
				now,
			)

			if err := repo.Add(ctx, testTask); err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			got, err := repo.Get(ctx, testTask.ID())
			if err != nil {
				t.Fatalf("failed to get inserted task: %v", err)
			}

			assertTaskEqual(t, testTask, got)
		})

		t.Run("adds task without assignees", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000011",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			testTask := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000012",
				projectID,
				nil,
				now,
			)

			if err := repo.Add(ctx, testTask); err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			got, err := repo.Get(ctx, testTask.ID())
			if err != nil {
				t.Fatalf("failed to get inserted task: %v", err)
			}

			assertTaskEqual(t, testTask, got)
		})

		t.Run("returns duplicate task ID error", func(t *testing.T) {
			firstProjectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000021",
			)

			secondProjectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000023",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			first := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000022",
				firstProjectID,
				nil,
				now,
			)

			if err := repo.Add(ctx, first); err != nil {
				t.Fatalf("failed to add first task: %v", err)
			}

			second := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000022",
				secondProjectID,
				nil,
				now.Add(time.Second),
			)

			err := repo.Add(ctx, second)
			if err == nil {
				t.Fatal("expected duplicate task ID error")
			}

			if !contains(err.Error(), "duplicate task id") {
				t.Fatalf(
					"expected duplicate task ID error, got %v",
					err,
				)
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("returns task with assignees", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000031",
			)

			userID1 := newTestTaskUser(
				t,
				db,
				"10000000-0000-0000-0000-000000000032",
			)
			userID2 := newTestTaskUser(
				t,
				db,
				"10000000-0000-0000-0000-000000000033",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			testTask := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000034",
				projectID,
				[]iam.UserID{userID1, userID2},
				now,
			)

			if err := repo.Add(ctx, testTask); err != nil {
				t.Fatalf("failed to add task: %v", err)
			}

			got, err := repo.Get(ctx, testTask.ID())
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			assertTaskEqual(t, testTask, got)
		})

		t.Run("returns not found", func(t *testing.T) {
			taskID, err := task.NewTaskID(
				"10000000-0000-0000-0000-000000000041",
			)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.Get(ctx, taskID)

			if got != nil {
				t.Fatal("expected nil task")
			}

			if !errors.Is(err, task.ErrTaskNotFound) {
				t.Fatalf(
					"expected ErrTaskNotFound, got %v",
					err,
				)
			}
		})
	})

	t.Run("ListCompletedPerDay", func(t *testing.T) {
		t.Run("returns completed counts for the last 14 days", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000051",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)
			completedAt := time.Date(
				now.Year(),
				now.Month(),
				now.Day()-1,
				12, 0, 0, 0,
				time.UTC,
			)

			completed := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000052",
				projectID,
				nil,
				completedAt,
			)

			if err := completed.MoveToInProgress(completedAt); err != nil {
				t.Fatalf("failed to move task to in progress: %v", err)
			}

			if err := completed.MoveToCompleted(completedAt); err != nil {
				t.Fatalf("failed to complete task: %v", err)
			}

			if err := repo.Add(ctx, completed); err != nil {
				t.Fatalf("failed to add completed task: %v", err)
			}

			otherProjectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000053",
			)

			other := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000054",
				otherProjectID,
				nil,
				now,
			)

			if err := other.MoveToInProgress(now); err != nil {
				t.Fatalf("failed to move other task to in progress: %v", err)
			}

			if err := other.MoveToCompleted(now); err != nil {
				t.Fatalf("failed to complete other task: %v", err)
			}

			if err := repo.Add(ctx, other); err != nil {
				t.Fatalf("failed to add other task: %v", err)
			}

			metrics, err := repo.ListCompletedPerDay(ctx, projectID)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if len(metrics) != 14 {
				t.Fatalf(
					"expected 14 metrics, got %d",
					len(metrics),
				)
			}

			expectedDay := completedAt.Format("2006-01-02")
			foundCompleted := false

			for _, metric := range metrics {
				if metric.Key() != expectedDay {
					continue
				}

				foundCompleted = true

				if metric.Value() != 1 {
					t.Fatalf(
						"expected 1 completed task on %s, got %d",
						expectedDay,
						metric.Value(),
					)
				}
			}

			if !foundCompleted {
				t.Fatalf(
					"expected metric for %s",
					expectedDay,
				)
			}
		})
	})

	t.Run("GetStatsByProject", func(t *testing.T) {
		t.Run("returns total and completed counts", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000061",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			task1 := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000062",
				projectID,
				nil,
				now,
			)

			task2 := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000063",
				projectID,
				nil,
				now.Add(time.Second),
			)

			if err := task2.MoveToInProgress(now.Add(time.Second)); err != nil {
				t.Fatalf("failed to move task to in progress: %v", err)
			}

			if err := task2.MoveToCompleted(now.Add(2 * time.Second)); err != nil {
				t.Fatalf("failed to complete task: %v", err)
			}

			if err := repo.Add(ctx, task1); err != nil {
				t.Fatalf("failed to add task1: %v", err)
			}

			if err := repo.Add(ctx, task2); err != nil {
				t.Fatalf("failed to add task2: %v", err)
			}

			stats, err := repo.GetStatsByProject(ctx, projectID)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if stats.Total() != 2 {
				t.Fatalf("expected total 2, got %d", stats.Total())
			}

			if stats.Completed() != 1 {
				t.Fatalf(
					"expected completed 1, got %d",
					stats.Completed(),
				)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("lists tasks with pagination", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000071",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			task1 := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000072",
				projectID,
				nil,
				now,
			)

			task2 := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000073",
				projectID,
				nil,
				now.Add(time.Second),
			)

			task3 := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000074",
				projectID,
				nil,
				now.Add(2*time.Second),
			)

			for _, testTask := range []*task.Task{
				task1,
				task2,
				task3,
			} {
				if err := repo.Add(ctx, testTask); err != nil {
					t.Fatalf("failed to add task: %v", err)
				}
			}

			page, err := common.NewPage(2, 0)
			if err != nil {
				t.Fatal(err)
			}

			tasks, err := repo.List(
				ctx,
				task.TaskFilter{
					ProjectID: projectID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if len(tasks) != 2 {
				t.Fatalf("expected 2 tasks, got %d", len(tasks))
			}

			if tasks[0].ID() != task3.ID() {
				t.Fatalf("expected newest task first")
			}

			if tasks[1].ID() != task2.ID() {
				t.Fatalf("expected second newest task")
			}
		})

		t.Run("loads assignees", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000075",
			)

			userID1 := newTestTaskUser(
				t,
				db,
				"10000000-0000-0000-0000-000000000076",
			)
			userID2 := newTestTaskUser(
				t,
				db,
				"10000000-0000-0000-0000-000000000077",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			testTask := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000078",
				projectID,
				[]iam.UserID{userID1, userID2},
				now,
			)

			if err := repo.Add(ctx, testTask); err != nil {
				t.Fatalf("failed to add task: %v", err)
			}

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			tasks, err := repo.List(
				ctx,
				task.TaskFilter{
					ProjectID: projectID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if len(tasks) != 1 {
				t.Fatalf("expected 1 task, got %d", len(tasks))
			}

			assertAssigneesEqual(
				t,
				testTask.Assignees(),
				tasks[0].Assignees(),
			)
		})

		t.Run("filters by keyword", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000081",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			task1 := newTestTaskWithName(
				t,
				"10000000-0000-0000-0000-000000000082",
				projectID,
				"Alpha task",
				"Description one",
				now,
			)

			task2 := newTestTaskWithName(
				t,
				"10000000-0000-0000-0000-000000000083",
				projectID,
				"Beta task",
				"Alpha description",
				now.Add(time.Second),
			)

			task3 := newTestTaskWithName(
				t,
				"10000000-0000-0000-0000-000000000084",
				projectID,
				"Gamma task",
				"Other description",
				now.Add(2*time.Second),
			)

			for _, testTask := range []*task.Task{
				task1,
				task2,
				task3,
			} {
				if err := repo.Add(ctx, testTask); err != nil {
					t.Fatalf("failed to add task: %v", err)
				}
			}

			keyword := "alpha"

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			tasks, err := repo.List(
				ctx,
				task.TaskFilter{
					ProjectID: projectID,
					Keyword:   &keyword,
				},
				page,
			)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if len(tasks) != 2 {
				t.Fatalf("expected 2 tasks, got %d", len(tasks))
			}
		})

		t.Run("filters by status", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000091",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			backlog := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000092",
				projectID,
				nil,
				now,
			)

			inProgress := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000093",
				projectID,
				nil,
				now.Add(time.Second),
			)

			if err := inProgress.MoveToInProgress(now.Add(time.Second)); err != nil {
				t.Fatalf("failed to move task to in progress: %v", err)
			}

			if err := repo.Add(ctx, backlog); err != nil {
				t.Fatalf("failed to add backlog task: %v", err)
			}

			if err := repo.Add(ctx, inProgress); err != nil {
				t.Fatalf("failed to add in-progress task: %v", err)
			}

			status := task.StatusInProgress

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			tasks, err := repo.List(
				ctx,
				task.TaskFilter{
					ProjectID: projectID,
					Status:    &status,
				},
				page,
			)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if len(tasks) != 1 {
				t.Fatalf("expected 1 task, got %d", len(tasks))
			}

			if tasks[0].ID() != inProgress.ID() {
				t.Fatal("expected in-progress task")
			}
		})

		t.Run("filters by priority", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-0000000000a1",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			low := newTestTaskWithPriority(
				t,
				"10000000-0000-0000-0000-0000000000a2",
				projectID,
				task.PriorityLow,
				now,
			)

			high := newTestTaskWithPriority(
				t,
				"10000000-0000-0000-0000-0000000000a3",
				projectID,
				task.PriorityHigh,
				now.Add(time.Second),
			)

			if err := repo.Add(ctx, low); err != nil {
				t.Fatalf("failed to add low-priority task: %v", err)
			}

			if err := repo.Add(ctx, high); err != nil {
				t.Fatalf("failed to add high-priority task: %v", err)
			}

			priority := task.PriorityHigh

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			tasks, err := repo.List(
				ctx,
				task.TaskFilter{
					ProjectID: projectID,
					Priority:  &priority,
				},
				page,
			)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if len(tasks) != 1 {
				t.Fatalf("expected 1 task, got %d", len(tasks))
			}

			if tasks[0].ID() != high.ID() {
				t.Fatal("expected high-priority task")
			}
		})
	})

	t.Run("ListStatsByProjects", func(t *testing.T) {
		t.Run("returns stats for multiple projects", func(t *testing.T) {
			projectID1 := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-0000000000b1",
			)

			projectID2 := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-0000000000b2",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			task1 := newTestTask(
				t,
				"10000000-0000-0000-0000-0000000000b3",
				projectID1,
				nil,
				now,
			)

			task2 := newTestTask(
				t,
				"10000000-0000-0000-0000-0000000000b4",
				projectID1,
				nil,
				now.Add(time.Second),
			)

			if err := task2.MoveToInProgress(now.Add(time.Second)); err != nil {
				t.Fatalf("failed to move task to in progress: %v", err)
			}

			if err := task2.MoveToCompleted(now.Add(2 * time.Second)); err != nil {
				t.Fatalf("failed to complete task: %v", err)
			}

			task3 := newTestTask(
				t,
				"10000000-0000-0000-0000-0000000000b5",
				projectID2,
				nil,
				now,
			)

			for _, testTask := range []*task.Task{
				task1,
				task2,
				task3,
			} {
				if err := repo.Add(ctx, testTask); err != nil {
					t.Fatalf("failed to add task: %v", err)
				}
			}

			stats, err := repo.ListStatsByProjects(
				ctx,
				[]project.ProjectID{projectID1, projectID2},
			)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			stats1, ok := stats[projectID1]
			if !ok {
				t.Fatal("expected stats for project 1")
			}

			if stats1.Total() != 2 || stats1.Completed() != 1 {
				t.Fatalf(
					"expected project 1 stats 2/1, got %d/%d",
					stats1.Total(),
					stats1.Completed(),
				)
			}

			stats2, ok := stats[projectID2]
			if !ok {
				t.Fatal("expected stats for project 2")
			}

			if stats2.Total() != 1 || stats2.Completed() != 0 {
				t.Fatalf(
					"expected project 2 stats 1/0, got %d/%d",
					stats2.Total(),
					stats2.Completed(),
				)
			}
		})

		t.Run("returns empty map for no projects", func(t *testing.T) {
			stats, err := repo.ListStatsByProjects(ctx, nil)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if stats == nil {
				t.Fatal("expected non-nil empty map")
			}

			if len(stats) != 0 {
				t.Fatalf(
					"expected empty map, got %d entries",
					len(stats),
				)
			}
		})
	})

	t.Run("Count", func(t *testing.T) {
		t.Run("counts tasks using filters", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-0000000000c1",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			match := newTestTaskWithName(
				t,
				"10000000-0000-0000-0000-0000000000c2",
				projectID,
				"Important task",
				"Important description",
				now,
			)

			other := newTestTaskWithName(
				t,
				"10000000-0000-0000-0000-0000000000c3",
				projectID,
				"Other task",
				"Other description",
				now.Add(time.Second),
			)

			if err := repo.Add(ctx, match); err != nil {
				t.Fatalf("failed to add match: %v", err)
			}

			if err := repo.Add(ctx, other); err != nil {
				t.Fatalf("failed to add other task: %v", err)
			}

			keyword := "important"

			count, err := repo.Count(
				ctx,
				task.TaskFilter{
					ProjectID: projectID,
					Keyword:   &keyword,
				},
			)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if count != 1 {
				t.Fatalf("expected count 1, got %d", count)
			}
		})
	})

	t.Run("Save", func(t *testing.T) {
		t.Run("updates task and replaces assignees", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-0000000000d1",
			)

			userID1 := newTestTaskUser(
				t,
				db,
				"10000000-0000-0000-0000-0000000000d2",
			)

			userID2 := newTestTaskUser(
				t,
				db,
				"10000000-0000-0000-0000-0000000000d3",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)
			updatedAt := now.Add(time.Second)

			testTask := newTestTask(
				t,
				"10000000-0000-0000-0000-0000000000d4",
				projectID,
				[]iam.UserID{userID1},
				now,
			)

			if err := repo.Add(ctx, testTask); err != nil {
				t.Fatalf("failed to add task: %v", err)
			}

			updatedName, err := task.NewName("Updated task")
			if err != nil {
				t.Fatal(err)
			}

			testTask = task.RestoreTask(
				testTask.ID(),
				testTask.ProjectID(),
				testTask.Priority(),
				testTask.Status(),
				updatedName,
				"Updated description",
				testTask.EstimatedMinutes(),
				[]iam.UserID{userID2},
				testTask.Version(),
				testTask.CreatedAt(),
				updatedAt,
			)

			if err := repo.Save(ctx, testTask); err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			got, err := repo.Get(ctx, testTask.ID())
			if err != nil {
				t.Fatalf("failed to retrieve saved task: %v", err)
			}

			if got.Name().String() != "Updated task" {
				t.Fatalf(
					"expected updated name, got %q",
					got.Name().String(),
				)
			}

			if got.Description() != "Updated description" {
				t.Fatalf(
					"expected updated description, got %q",
					got.Description(),
				)
			}

			if got.Version() != testTask.Version()+1 {
				t.Fatalf(
					"expected version %d, got %d",
					testTask.Version()+1,
					got.Version(),
				)
			}

			if !got.UpdatedAt().Equal(updatedAt) {
				t.Fatalf(
					"expected updated_at %v, got %v",
					updatedAt,
					got.UpdatedAt(),
				)
			}

			assertAssigneesEqual(
				t,
				[]iam.UserID{userID2},
				got.Assignees(),
			)
		})

		t.Run("returns concurrent modification error", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-0000000000e1",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			testTask := newTestTask(
				t,
				"10000000-0000-0000-0000-0000000000e2",
				projectID,
				nil,
				now,
			)

			if err := repo.Add(ctx, testTask); err != nil {
				t.Fatalf("failed to add task: %v", err)
			}

			stale := task.RestoreTask(
				testTask.ID(),
				testTask.ProjectID(),
				testTask.Priority(),
				testTask.Status(),
				testTask.Name(),
				testTask.Description(),
				testTask.EstimatedMinutes(),
				testTask.Assignees(),
				testTask.Version()+1,
				testTask.CreatedAt(),
				testTask.UpdatedAt(),
			)

			err := repo.Save(ctx, stale)
			if err == nil {
				t.Fatal("expected concurrent modification error")
			}

			if !errors.Is(
				err,
				task.ErrTaskConcurrentModification,
			) {
				t.Fatalf(
					"expected ErrTaskConcurrentModification, got %v",
					err,
				)
			}
		})

		t.Run("updates task with no assignees", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-0000000000f1",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			testTask := newTestTask(
				t,
				"10000000-0000-0000-0000-0000000000f2",
				projectID,
				nil,
				now,
			)

			if err := repo.Add(ctx, testTask); err != nil {
				t.Fatalf("failed to add task: %v", err)
			}

			updatedName, err := task.NewName("Updated without assignees")
			if err != nil {
				t.Fatal(err)
			}

			testTask = task.RestoreTask(
				testTask.ID(),
				testTask.ProjectID(),
				testTask.Priority(),
				testTask.Status(),
				updatedName,
				"Updated description",
				testTask.EstimatedMinutes(),
				nil,
				testTask.Version(),
				testTask.CreatedAt(),
				now.Add(time.Second),
			)

			if err := repo.Save(ctx, testTask); err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			got, err := repo.Get(ctx, testTask.ID())
			if err != nil {
				t.Fatalf("failed to retrieve saved task: %v", err)
			}

			if got.Name().String() != "Updated without assignees" {
				t.Fatalf(
					"expected updated name, got %q",
					got.Name().String(),
				)
			}

			if len(got.Assignees()) != 0 {
				t.Fatalf(
					"expected no assignees, got %d",
					len(got.Assignees()),
				)
			}
		})
	})

	t.Run("Remove", func(t *testing.T) {
		t.Run("removes task", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000101",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			testTask := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000102",
				projectID,
				nil,
				now,
			)

			if err := repo.Add(ctx, testTask); err != nil {
				t.Fatalf("failed to add task: %v", err)
			}

			if err := repo.Remove(ctx, testTask); err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			_, err := repo.Get(ctx, testTask.ID())
			if !errors.Is(err, task.ErrTaskNotFound) {
				t.Fatalf(
					"expected ErrTaskNotFound after removal, got %v",
					err,
				)
			}
		})

		t.Run("returns concurrent modification error", func(t *testing.T) {
			projectID := newTestTaskProject(
				t,
				db,
				"10000000-0000-0000-0000-000000000111",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			testTask := newTestTask(
				t,
				"10000000-0000-0000-0000-000000000112",
				projectID,
				nil,
				now,
			)

			if err := repo.Add(ctx, testTask); err != nil {
				t.Fatalf("failed to add task: %v", err)
			}

			stale := task.RestoreTask(
				testTask.ID(),
				testTask.ProjectID(),
				testTask.Priority(),
				testTask.Status(),
				testTask.Name(),
				testTask.Description(),
				testTask.EstimatedMinutes(),
				testTask.Assignees(),
				testTask.Version()+1,
				testTask.CreatedAt(),
				testTask.UpdatedAt(),
			)

			err := repo.Remove(ctx, stale)
			if err == nil {
				t.Fatal("expected concurrent modification error")
			}

			if !errors.Is(
				err,
				task.ErrTaskConcurrentModification,
			) {
				t.Fatalf(
					"expected ErrTaskConcurrentModification, got %v",
					err,
				)
			}
		})
	})
}

func newTestTaskUser(
	t *testing.T,
	db *sql.DB,
	id string,
) iam.UserID {
	t.Helper()

	_, err := db.ExecContext(
		context.Background(),
		`INSERT INTO users(
			id,
			first_name,
			last_name,
			email,
			role
		) VALUES (?, ?, ?, ?, ?)`,
		id,
		"Test",
		"User",
		id+"@example.com",
		"client",
	)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	userID, err := iam.NewUserID(id)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}

func newTestTaskProject(
	t *testing.T,
	db *sql.DB,
	id string,
) project.ProjectID {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)

	_, err := db.ExecContext(
		context.Background(),
		`INSERT INTO projects(
			id,
			name,
			status,
			created_at,
			updated_at,
			version
		) VALUES (?, ?, ?, ?, ?, ?)`,
		id,
		"Test Project "+id,
		"active",
		now,
		now,
		1,
	)
	if err != nil {
		t.Fatalf("failed to insert test project: %v", err)
	}

	projectID, err := project.NewProjectID(id)
	if err != nil {
		t.Fatal(err)
	}

	return projectID
}

func newTestTask(
	t *testing.T,
	id string,
	projectID project.ProjectID,
	assigneeIDs []iam.UserID,
	updatedAt time.Time,
) *task.Task {
	t.Helper()

	taskID, err := task.NewTaskID(id)
	if err != nil {
		t.Fatalf("failed to create task ID: %v", err)
	}

	name, err := task.NewName("Test Task " + id)
	if err != nil {
		t.Fatalf("failed to create task name: %v", err)
	}

	testTask, err := task.NewTask(
		taskID,
		projectID,
		task.PriorityMedium,
		task.StatusBacklog,
		name,
		"Test task description",
		task.Minutes(60),
		assigneeIDs,
		updatedAt,
	)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	return testTask
}
func newTestTaskWithName(
	t *testing.T,
	id string,
	projectID project.ProjectID,
	name string,
	description string,
	now time.Time,
) *task.Task {
	t.Helper()

	taskID, err := task.NewTaskID(id)
	if err != nil {
		t.Fatal(err)
	}

	priority, err := task.NewPriority(task.PriorityMedium.String())
	if err != nil {
		t.Fatal(err)
	}

	status, err := task.NewStatus(task.StatusBacklog.String())
	if err != nil {
		t.Fatal(err)
	}

	nameVO, err := task.NewName(name)
	if err != nil {
		t.Fatal(err)
	}

	minutes, err := task.NewMinutes(60)
	if err != nil {
		t.Fatal(err)
	}

	testTask, err := task.NewTask(
		taskID,
		projectID,
		priority,
		status,
		nameVO,
		description,
		minutes,
		nil,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}

	return testTask
}

func newTestTaskWithPriority(
	t *testing.T,
	id string,
	projectID project.ProjectID,
	priority task.Priority,
	now time.Time,
) *task.Task {
	t.Helper()

	taskID, err := task.NewTaskID(id)
	if err != nil {
		t.Fatal(err)
	}

	status, err := task.NewStatus(task.StatusBacklog.String())
	if err != nil {
		t.Fatal(err)
	}

	name, err := task.NewName("Test task " + id)
	if err != nil {
		t.Fatal(err)
	}

	minutes, err := task.NewMinutes(60)
	if err != nil {
		t.Fatal(err)
	}

	testTask, err := task.NewTask(
		taskID,
		projectID,
		priority,
		status,
		name,
		"Test description",
		minutes,
		nil,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}

	return testTask
}

func assertTaskEqual(
	t *testing.T,
	expected *task.Task,
	actual *task.Task,
) {
	t.Helper()

	if actual == nil {
		t.Fatal("expected task, got nil")
	}

	if actual.ID() != expected.ID() {
		t.Fatalf(
			"expected ID %q, got %q",
			expected.ID(),
			actual.ID(),
		)
	}

	if actual.ProjectID() != expected.ProjectID() {
		t.Fatalf(
			"expected project ID %q, got %q",
			expected.ProjectID(),
			actual.ProjectID(),
		)
	}

	if actual.Priority() != expected.Priority() {
		t.Fatalf(
			"expected priority %q, got %q",
			expected.Priority(),
			actual.Priority(),
		)
	}

	if actual.Status() != expected.Status() {
		t.Fatalf(
			"expected status %q, got %q",
			expected.Status(),
			actual.Status(),
		)
	}

	if actual.Name() != expected.Name() {
		t.Fatalf(
			"expected name %q, got %q",
			expected.Name(),
			actual.Name(),
		)
	}

	if actual.Description() != expected.Description() {
		t.Fatalf(
			"expected description %q, got %q",
			expected.Description(),
			actual.Description(),
		)
	}

	if actual.EstimatedMinutes() != expected.EstimatedMinutes() {
		t.Fatalf(
			"expected estimated minutes %d, got %d",
			expected.EstimatedMinutes(),
			actual.EstimatedMinutes(),
		)
	}

	if actual.Version() != expected.Version() {
		t.Fatalf(
			"expected version %d, got %d",
			expected.Version(),
			actual.Version(),
		)
	}

	if !actual.CreatedAt().Equal(expected.CreatedAt()) {
		t.Fatalf(
			"expected created_at %v, got %v",
			expected.CreatedAt(),
			actual.CreatedAt(),
		)
	}

	if !actual.UpdatedAt().Equal(expected.UpdatedAt()) {
		t.Fatalf(
			"expected updated_at %v, got %v",
			expected.UpdatedAt(),
			actual.UpdatedAt(),
		)
	}

	assertAssigneesEqual(
		t,
		expected.Assignees(),
		actual.Assignees(),
	)
}

func assertAssigneesEqual(
	t *testing.T,
	expected []iam.UserID,
	actual []iam.UserID,
) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf(
			"expected %d assignees, got %d",
			len(expected),
			len(actual),
		)
	}

	expectedSet := make(map[iam.UserID]struct{}, len(expected))
	for _, id := range expected {
		expectedSet[id] = struct{}{}
	}

	actualSet := make(map[iam.UserID]struct{}, len(actual))
	for _, id := range actual {
		actualSet[id] = struct{}{}
	}

	for id := range expectedSet {
		if _, ok := actualSet[id]; !ok {
			t.Fatalf("expected assignee %q", id)
		}
	}

	for id := range actualSet {
		if _, ok := expectedSet[id]; !ok {
			t.Fatalf("unexpected assignee %q", id)
		}
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
