package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

func TestProjectRepository(t *testing.T) {
	db := newTestPostgres(t)
	repo := NewProjectRepository(db)
	ctx := context.Background()

	user1ID := newTestProjectUser(
		t,
		db,
		"30000000-0000-0000-0000-000000000001",
	)

	user2ID := newTestProjectUser(
		t,
		db,
		"30000000-0000-0000-0000-000000000002",
	)

	user3ID := newTestProjectUser(
		t,
		db,
		"30000000-0000-0000-0000-000000000003",
	)

	user4ID := newTestProjectUser(
		t,
		db,
		"30000000-0000-0000-0000-000000000004",
	)

	t.Run("Add", func(t *testing.T) {
		t.Run("adds project with members", func(t *testing.T) {
			p := newTestProject(
				t,
				"31000000-0000-0000-0000-000000000001",
				"Project One",
				"started",
				[]iam.UserID{user1ID, user2ID},
				1,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, p); err != nil {
				t.Fatalf("failed to add project: %v", err)
			}

			got, err := repo.Get(ctx, p.ID())
			if err != nil {
				t.Fatalf("failed to get project: %v", err)
			}

			assertProjectEqual(t, p, got)
		})

		t.Run("returns error for duplicate project", func(t *testing.T) {
			p := newTestProject(
				t,
				"31000000-0000-0000-0000-000000000002",
				"Duplicate Project",
				"started",
				[]iam.UserID{user1ID},
				1,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, p); err != nil {
				t.Fatalf("failed to add initial project: %v", err)
			}

			err := repo.Add(ctx, p)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("returns project with members", func(t *testing.T) {
			p := newTestProject(
				t,
				"32000000-0000-0000-0000-000000000001",
				"Get Project",
				"started",
				[]iam.UserID{user1ID, user2ID},
				3,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, p); err != nil {
				t.Fatalf("failed to add project: %v", err)
			}

			got, err := repo.Get(ctx, p.ID())
			if err != nil {
				t.Fatalf("failed to get project: %v", err)
			}

			assertProjectEqual(t, p, got)
		})

		t.Run("returns project with no members", func(t *testing.T) {
			p := newTestProject(
				t,
				"32000000-0000-0000-0000-000000000002",
				"No Members",
				"started",
				nil,
				1,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			insertTestProjectDirectly(t, db, p)

			got, err := repo.Get(ctx, p.ID())
			if err != nil {
				t.Fatalf("failed to get project: %v", err)
			}

			assertProjectEqual(t, p, got)
		})

		t.Run("returns not found", func(t *testing.T) {
			id, err := project.NewProjectID(
				"32000000-0000-0000-0000-000000000010",
			)
			if err != nil {
				t.Fatal(err)
			}

			_, err = repo.Get(ctx, id)

			if !errors.Is(err, project.ErrProjectNotFound) {
				t.Fatalf(
					"expected ErrProjectNotFound, got %v",
					err,
				)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("lists projects for member", func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Microsecond)

			p1 := newTestProject(
				t,
				"33000000-0000-0000-0000-000000000001",
				"Member Project One",
				"started",
				[]iam.UserID{user4ID, user2ID},
				1,
				now,
			)

			p2 := newTestProject(
				t,
				"33000000-0000-0000-0000-000000000002",
				"Member Project Two",
				"started",
				[]iam.UserID{user4ID},
				1,
				now.Add(time.Second),
			)

			p3 := newTestProject(
				t,
				"33000000-0000-0000-0000-000000000003",
				"Other Member Project",
				"started",
				[]iam.UserID{user2ID},
				1,
				now.Add(2*time.Second),
			)

			insertTestProjectDirectly(t, db, p1)
			insertTestProjectDirectly(t, db, p2)
			insertTestProjectDirectly(t, db, p3)

			status, err := project.NewStatus("started")
			if err != nil {
				t.Fatal(err)
			}

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			projects, err := repo.List(
				ctx,
				project.Filter{
					MemberID: user4ID,
					Status:   &status,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list projects: %v", err)
			}

			if len(projects) != 2 {
				t.Fatalf("expected 2 projects, got %d", len(projects))
			}

			if projects[0].ID() != p2.ID() {
				t.Fatalf(
					"expected first project %s, got %s",
					p2.ID(),
					projects[0].ID(),
				)
			}

			if projects[1].ID() != p1.ID() {
				t.Fatalf(
					"expected second project %s, got %s",
					p1.ID(),
					projects[1].ID(),
				)
			}

			assertProjectMembers(
				t,
				projects[0],
				[]iam.UserID{user4ID},
			)

			assertProjectMembers(
				t,
				projects[1],
				[]iam.UserID{user4ID, user2ID},
			)
		})

		t.Run("filters by keyword", func(t *testing.T) {
			p1 := newTestProject(
				t,
				"33100000-0000-0000-0000-000000000001",
				"Alpha Website",
				"started",
				[]iam.UserID{user1ID},
				1,
				time.Now().UTC(),
			)

			p2 := newTestProject(
				t,
				"33100000-0000-0000-0000-000000000002",
				"Beta Website",
				"started",
				[]iam.UserID{user1ID},
				1,
				time.Now().UTC().Add(time.Second),
			)

			p3 := newTestProject(
				t,
				"33100000-0000-0000-0000-000000000003",
				"Mobile Application",
				"started",
				[]iam.UserID{user1ID},
				1,
				time.Now().UTC().Add(2*time.Second),
			)

			insertTestProjectDirectly(t, db, p1)
			insertTestProjectDirectly(t, db, p2)
			insertTestProjectDirectly(t, db, p3)

			keyword := "website"

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			projects, err := repo.List(
				ctx,
				project.Filter{
					MemberID: user1ID,
					Keyword:  &keyword,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list projects: %v", err)
			}

			if len(projects) != 2 {
				t.Fatalf("expected 2 projects, got %d", len(projects))
			}

			if projects[0].ID() != p2.ID() {
				t.Fatalf("expected %s, got %s", p2.ID(), projects[0].ID())
			}

			if projects[1].ID() != p1.ID() {
				t.Fatalf("expected %s, got %s", p1.ID(), projects[1].ID())
			}
		})

		t.Run("filters by status", func(t *testing.T) {
			filterUserID := newTestProjectUser(
				t,
				db,
				"30000000-0000-0000-0000-000000000005",
			)

			started := newTestProject(
				t,
				"33200000-0000-0000-0000-000000000005",
				"Started Project",
				"started",
				[]iam.UserID{filterUserID},
				1,
				time.Now().UTC(),
			)

			completed := newTestProject(
				t,
				"33200000-0000-0000-0000-000000000006",
				"Completed Project",
				"completed",
				[]iam.UserID{filterUserID},
				1,
				time.Now().UTC().Add(time.Second),
			)

			insertTestProjectDirectly(t, db, started)
			insertTestProjectDirectly(t, db, completed)

			status, err := project.NewStatus("started")
			if err != nil {
				t.Fatal(err)
			}

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			projects, err := repo.List(
				ctx,
				project.Filter{
					MemberID: filterUserID,
					Status:   &status,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list projects: %v", err)
			}

			if len(projects) != 1 {
				t.Fatalf("expected 1 project, got %d", len(projects))
			}

			if projects[0].ID() != started.ID() {
				t.Fatalf(
					"expected %s, got %s",
					started.ID(),
					projects[0].ID(),
				)
			}
		})

		t.Run("applies pagination", func(t *testing.T) {
			filterUserID := newTestProjectUser(
				t,
				db,
				"30000000-0000-0000-0000-000000000006",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			p1 := newTestProject(
				t,
				"33300000-0000-0000-0000-000000000001",
				"Pagination One",
				"started",
				[]iam.UserID{filterUserID},
				1,
				now,
			)

			p2 := newTestProject(
				t,
				"33300000-0000-0000-0000-000000000002",
				"Pagination Two",
				"started",
				[]iam.UserID{filterUserID},
				1,
				now.Add(time.Second),
			)

			p3 := newTestProject(
				t,
				"33300000-0000-0000-0000-000000000003",
				"Pagination Three",
				"started",
				[]iam.UserID{filterUserID},
				1,
				now.Add(2*time.Second),
			)

			insertTestProjectDirectly(t, db, p1)
			insertTestProjectDirectly(t, db, p2)
			insertTestProjectDirectly(t, db, p3)

			page, err := common.NewPage(1, 1)
			if err != nil {
				t.Fatal(err)
			}

			projects, err := repo.List(
				ctx,
				project.Filter{MemberID: filterUserID},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list projects: %v", err)
			}

			if len(projects) != 1 {
				t.Fatalf("expected 1 project, got %d", len(projects))
			}

			if projects[0].ID() != p2.ID() {
				t.Fatalf(
					"expected %s, got %s",
					p2.ID(),
					projects[0].ID(),
				)
			}
		})

		t.Run("returns nil for no matches", func(t *testing.T) {
			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			projects, err := repo.List(
				ctx,
				project.Filter{MemberID: user3ID},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list projects: %v", err)
			}

			if projects != nil {
				t.Fatalf("expected nil, got %v", projects)
			}
		})
	})

	t.Run("ListByIDs", func(t *testing.T) {
		t.Run("lists requested projects with members", func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Microsecond)

			p1 := newTestProject(
				t,
				"34000000-0000-0000-0000-000000000001",
				"List By IDs One",
				"started",
				[]iam.UserID{user1ID, user2ID},
				1,
				now,
			)

			p2 := newTestProject(
				t,
				"34000000-0000-0000-0000-000000000002",
				"List By IDs Two",
				"completed",
				[]iam.UserID{user2ID, user3ID},
				2,
				now.Add(time.Second),
			)

			p3 := newTestProject(
				t,
				"34000000-0000-0000-0000-000000000003",
				"Not Requested",
				"started",
				[]iam.UserID{user3ID},
				1,
				now.Add(2*time.Second),
			)

			insertTestProjectDirectly(t, db, p1)
			insertTestProjectDirectly(t, db, p2)
			insertTestProjectDirectly(t, db, p3)

			projects, err := repo.ListByIDs(
				ctx,
				[]project.ProjectID{p2.ID(), p1.ID()},
			)
			if err != nil {
				t.Fatalf("failed to list projects: %v", err)
			}

			if len(projects) != 2 {
				t.Fatalf("expected 2 projects, got %d", len(projects))
			}

			got := make(map[project.ProjectID]*project.Project)

			for _, p := range projects {
				got[p.ID()] = p
			}

			assertProjectEqual(t, p1, got[p1.ID()])
			assertProjectEqual(t, p2, got[p2.ID()])
		})

		t.Run("returns not found when a project is missing", func(t *testing.T) {
			p := newTestProject(
				t,
				"34100000-0000-0000-0000-000000000001",
				"Existing Project",
				"started",
				[]iam.UserID{user1ID},
				1,
				time.Now().UTC(),
			)

			insertTestProjectDirectly(t, db, p)

			missingID, err := project.NewProjectID(
				"34100000-0000-0000-0000-000000000002",
			)
			if err != nil {
				t.Fatal(err)
			}

			_, err = repo.ListByIDs(
				ctx,
				[]project.ProjectID{p.ID(), missingID},
			)

			if !errors.Is(err, project.ErrProjectNotFound) {
				t.Fatalf(
					"expected ErrProjectNotFound, got %v",
					err,
				)
			}
		})

		t.Run("returns nil for empty IDs", func(t *testing.T) {
			projects, err := repo.ListByIDs(
				ctx,
				nil,
			)
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if projects != nil {
				t.Fatalf("expected nil, got %v", projects)
			}
		})
	})

	t.Run("ListByMember", func(t *testing.T) {
		t.Run("lists projects for member", func(t *testing.T) {
			memberID := newTestProjectUser(
				t,
				db,
				"30000000-0000-0000-0000-000000000007",
			)

			otherMemberID := newTestProjectUser(
				t,
				db,
				"30000000-0000-0000-0000-000000000008",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			p1 := newTestProject(
				t,
				"35000000-0000-0000-0000-000000000001",
				"Member One",
				"started",
				[]iam.UserID{memberID, otherMemberID},
				1,
				now,
			)

			p2 := newTestProject(
				t,
				"35000000-0000-0000-0000-000000000002",
				"Member Two",
				"completed",
				[]iam.UserID{memberID},
				1,
				now.Add(time.Second),
			)

			p3 := newTestProject(
				t,
				"35000000-0000-0000-0000-000000000003",
				"Other Project",
				"started",
				[]iam.UserID{otherMemberID},
				1,
				now.Add(2*time.Second),
			)

			insertTestProjectDirectly(t, db, p1)
			insertTestProjectDirectly(t, db, p2)
			insertTestProjectDirectly(t, db, p3)

			projects, err := repo.ListByMember(ctx, memberID)
			if err != nil {
				t.Fatalf("failed to list projects: %v", err)
			}

			if len(projects) != 2 {
				t.Fatalf("expected 2 projects, got %d", len(projects))
			}

			if projects[0].ID() != p2.ID() {
				t.Fatalf(
					"expected first project %s, got %s",
					p2.ID(),
					projects[0].ID(),
				)
			}

			if projects[1].ID() != p1.ID() {
				t.Fatalf(
					"expected second project %s, got %s",
					p1.ID(),
					projects[1].ID(),
				)
			}

			assertProjectMembers(
				t,
				projects[0],
				[]iam.UserID{memberID},
			)

			assertProjectMembers(
				t,
				projects[1],
				[]iam.UserID{memberID, otherMemberID},
			)
		})

		t.Run("returns nil when member has no projects", func(t *testing.T) {
			memberID := newTestProjectUser(
				t,
				db,
				"30000000-0000-0000-0000-000000000009",
			)

			projects, err := repo.ListByMember(
				ctx,
				memberID,
			)
			if err != nil {
				t.Fatalf("failed to list projects: %v", err)
			}

			if projects != nil {
				t.Fatalf("expected nil, got %v", projects)
			}
		})
	})

	t.Run("ListCreatedPerDay", func(t *testing.T) {
		t.Run("returns twelve monthly metrics", func(t *testing.T) {
			metrics, err := repo.ListCreatedPerDay(
				ctx,
				user1ID,
			)
			if err != nil {
				t.Fatalf("failed to list metrics: %v", err)
			}

			if len(metrics) != 12 {
				t.Fatalf(
					"expected 12 metrics, got %d",
					len(metrics),
				)
			}
		})

		t.Run("counts projects for member", func(t *testing.T) {
			memberID := newTestProjectUser(
				t,
				db,
				"30000000-0000-0000-0000-000000000010",
			)

			otherMemberID := newTestProjectUser(
				t,
				db,
				"30000000-0000-0000-0000-000000000011",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			currentMonth := time.Date(
				now.Year(),
				now.Month(),
				10,
				12,
				0,
				0,
				0,
				time.UTC,
			)

			memberProject := newTestProject(
				t,
				"36000000-0000-0000-0000-000000000001",
				"Metric Member Project",
				"started",
				[]iam.UserID{memberID},
				1,
				currentMonth,
			)

			otherProject := newTestProject(
				t,
				"36000000-0000-0000-0000-000000000002",
				"Metric Other Project",
				"started",
				[]iam.UserID{otherMemberID},
				1,
				currentMonth.Add(time.Hour),
			)

			insertTestProjectDirectly(t, db, memberProject)
			insertTestProjectDirectly(t, db, otherProject)

			metrics, err := repo.ListCreatedPerDay(
				ctx,
				memberID,
			)
			if err != nil {
				t.Fatalf("failed to list metrics: %v", err)
			}

			currentKey := currentMonth.Format("2006-01")

			var currentCount int64
			var found bool

			for _, metric := range metrics {
				if metric.Key() == currentKey {
					currentCount = metric.Value()
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("expected metric for %s", currentKey)
			}

			if currentCount != 1 {
				t.Fatalf(
					"expected 1 project for member, got %d",
					currentCount,
				)
			}
		})
	})

	t.Run("Count", func(t *testing.T) {
		t.Run("counts projects for member", func(t *testing.T) {
			memberID := newTestProjectUser(
				t,
				db,
				"30000000-0000-0000-0000-000000000012",
			)

			otherMemberID := newTestProjectUser(
				t,
				db,
				"30000000-0000-0000-0000-000000000013",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			p1 := newTestProject(
				t,
				"37000000-0000-0000-0000-000000000001",
				"Count One",
				"started",
				[]iam.UserID{memberID},
				1,
				now,
			)

			p2 := newTestProject(
				t,
				"37000000-0000-0000-0000-000000000002",
				"Count Two",
				"started",
				[]iam.UserID{memberID},
				1,
				now.Add(time.Second),
			)

			p3 := newTestProject(
				t,
				"37000000-0000-0000-0000-000000000003",
				"Other Count",
				"started",
				[]iam.UserID{otherMemberID},
				1,
				now.Add(2*time.Second),
			)

			insertTestProjectDirectly(t, db, p1)
			insertTestProjectDirectly(t, db, p2)
			insertTestProjectDirectly(t, db, p3)

			count, err := repo.Count(
				ctx,
				project.Filter{MemberID: memberID},
			)
			if err != nil {
				t.Fatalf("failed to count projects: %v", err)
			}

			if count != 2 {
				t.Fatalf("expected 2, got %d", count)
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			memberID := newTestProjectUser(
				t,
				db,
				"30000000-0000-0000-0000-000000000014",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			p1 := newTestProject(
				t,
				"37100000-0000-0000-0000-000000000001",
				"Count Website",
				"started",
				[]iam.UserID{memberID},
				1,
				now,
			)

			p2 := newTestProject(
				t,
				"37100000-0000-0000-0000-000000000002",
				"Count Mobile",
				"started",
				[]iam.UserID{memberID},
				1,
				now.Add(time.Second),
			)

			insertTestProjectDirectly(t, db, p1)
			insertTestProjectDirectly(t, db, p2)

			keyword := "website"

			count, err := repo.Count(
				ctx,
				project.Filter{
					MemberID: memberID,
					Keyword:  &keyword,
				},
			)
			if err != nil {
				t.Fatalf("failed to count projects: %v", err)
			}

			if count != 1 {
				t.Fatalf("expected 1, got %d", count)
			}
		})

		t.Run("filters by status", func(t *testing.T) {
			memberID := newTestProjectUser(
				t,
				db,
				"30000000-0000-0000-0000-000000000015",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			started := newTestProject(
				t,
				"37200000-0000-0000-0000-000000000001",
				"Count Started",
				"started",
				[]iam.UserID{memberID},
				1,
				now,
			)

			completed := newTestProject(
				t,
				"37200000-0000-0000-0000-000000000002",
				"Count Completed",
				"completed",
				[]iam.UserID{memberID},
				1,
				now.Add(time.Second),
			)

			insertTestProjectDirectly(t, db, started)
			insertTestProjectDirectly(t, db, completed)

			status, err := project.NewStatus("started")
			if err != nil {
				t.Fatal(err)
			}

			count, err := repo.Count(
				ctx,
				project.Filter{
					MemberID: memberID,
					Status:   &status,
				},
			)
			if err != nil {
				t.Fatalf("failed to count projects: %v", err)
			}

			if count != 1 {
				t.Fatalf("expected 1, got %d", count)
			}
		})
	})

	t.Run("Save", func(t *testing.T) {
		t.Run("updates project and members", func(t *testing.T) {
			p := newTestProject(
				t,
				"38000000-0000-0000-0000-000000000001",
				"Original Name",
				"started",
				[]iam.UserID{user1ID, user2ID},
				1,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			insertTestProjectDirectly(t, db, p)

			name, err := project.NewName("Updated Name")
			if err != nil {
				t.Fatal(err)
			}

			status, err := project.NewStatus("completed")
			if err != nil {
				t.Fatal(err)
			}

			updatedAt := time.Now().UTC().Add(time.Second).Truncate(time.Microsecond)

			p = project.RestoreProject(
				p.ID(),
				name,
				status,
				[]iam.UserID{user1ID, user3ID},
				p.Version(),
				p.CreatedAt(),
				updatedAt,
			)

			if err := repo.Save(ctx, p); err != nil {
				t.Fatalf("failed to save project: %v", err)
			}

			got, err := repo.Get(ctx, p.ID())
			if err != nil {
				t.Fatalf("failed to get project: %v", err)
			}

			if got.Name() != p.Name() {
				t.Fatalf(
					"expected name %s, got %s",
					p.Name(),
					got.Name(),
				)
			}

			if got.Status() != p.Status() {
				t.Fatalf(
					"expected status %s, got %s",
					p.Status(),
					got.Status(),
				)
			}

			if got.Version() != p.Version()+1 {
				t.Fatalf(
					"expected version %d, got %d",
					p.Version()+1,
					got.Version(),
				)
			}

			assertProjectMembers(
				t,
				got,
				[]iam.UserID{user1ID, user3ID},
			)
		})

		t.Run("removes all members", func(t *testing.T) {
			p := newTestProject(
				t,
				"38100000-0000-0000-0000-000000000001",
				"Remove Members",
				"started",
				[]iam.UserID{user1ID, user2ID},
				1,
				time.Now().UTC(),
			)

			insertTestProjectDirectly(t, db, p)

			p = project.RestoreProject(
				p.ID(),
				p.Name(),
				p.Status(),
				nil,
				p.Version(),
				p.CreatedAt(),
				time.Now().UTC().Add(time.Second),
			)

			if err := repo.Save(ctx, p); err != nil {
				t.Fatalf("failed to save project: %v", err)
			}

			got, err := repo.Get(ctx, p.ID())
			if err != nil {
				t.Fatalf("failed to get project: %v", err)
			}

			if len(got.Members()) != 0 {
				t.Fatalf(
					"expected 0 members, got %d",
					len(got.Members()),
				)
			}
		})

		t.Run("returns concurrent modification", func(t *testing.T) {
			p := newTestProject(
				t,
				"38200000-0000-0000-0000-000000000001",
				"Concurrent Project",
				"started",
				[]iam.UserID{user1ID},
				2,
				time.Now().UTC(),
			)

			insertTestProjectDirectly(t, db, p)

			name, err := project.NewName("Changed")
			if err != nil {
				t.Fatal(err)
			}

			p = project.RestoreProject(
				p.ID(),
				name,
				p.Status(),
				p.Members(),
				1,
				p.CreatedAt(),
				time.Now().UTC().Add(time.Second),
			)

			err = repo.Save(ctx, p)

			if !errors.Is(
				err,
				project.ErrProjectConcurrentModification,
			) {
				t.Fatalf(
					"expected ErrProjectConcurrentModification, got %v",
					err,
				)
			}
		})
	})

	t.Run("SaveAll", func(t *testing.T) {
		t.Run("saves all projects", func(t *testing.T) {
			p1 := newTestProject(
				t,
				"39000000-0000-0000-0000-000000000001",
				"Save All One",
				"started",
				[]iam.UserID{user1ID},
				1,
				time.Now().UTC(),
			)

			p2 := newTestProject(
				t,
				"39000000-0000-0000-0000-000000000002",
				"Save All Two",
				"started",
				[]iam.UserID{user2ID},
				1,
				time.Now().UTC().Add(time.Second),
			)

			insertTestProjectDirectly(t, db, p1)
			insertTestProjectDirectly(t, db, p2)

			name1, err := project.NewName("Updated One")
			if err != nil {
				t.Fatal(err)
			}

			name2, err := project.NewName("Updated Two")
			if err != nil {
				t.Fatal(err)
			}

			p1 = project.RestoreProject(
				p1.ID(),
				name1,
				p1.Status(),
				p1.Members(),
				p1.Version(),
				p1.CreatedAt(),
				time.Now().UTC().Add(time.Second),
			)

			p2 = project.RestoreProject(
				p2.ID(),
				name2,
				p2.Status(),
				p2.Members(),
				p2.Version(),
				p2.CreatedAt(),
				time.Now().UTC().Add(2*time.Second),
			)

			if err := repo.SaveAll(ctx, []*project.Project{p1, p2}); err != nil {
				t.Fatalf("failed to save all projects: %v", err)
			}

			got1, err := repo.Get(ctx, p1.ID())
			if err != nil {
				t.Fatalf("failed to get first project: %v", err)
			}

			got2, err := repo.Get(ctx, p2.ID())
			if err != nil {
				t.Fatalf("failed to get second project: %v", err)
			}

			if got1.Name() != p1.Name() {
				t.Fatalf(
					"expected %s, got %s",
					p1.Name(),
					got1.Name(),
				)
			}

			if got2.Name() != p2.Name() {
				t.Fatalf(
					"expected %s, got %s",
					p2.Name(),
					got2.Name(),
				)
			}

			if got1.Version() != 2 {
				t.Fatalf("expected version 2, got %d", got1.Version())
			}

			if got2.Version() != 2 {
				t.Fatalf("expected version 2, got %d", got2.Version())
			}
		})
	})

	t.Run("Remove", func(t *testing.T) {
		t.Run("removes project", func(t *testing.T) {
			p := newTestProject(
				t,
				"40000000-0000-0000-0000-000000000001",
				"Remove Project",
				"started",
				[]iam.UserID{user1ID, user2ID},
				1,
				time.Now().UTC(),
			)

			insertTestProjectDirectly(t, db, p)

			if err := repo.Remove(ctx, p); err != nil {
				t.Fatalf("failed to remove project: %v", err)
			}

			_, err := repo.Get(ctx, p.ID())

			if !errors.Is(err, project.ErrProjectNotFound) {
				t.Fatalf(
					"expected ErrProjectNotFound, got %v",
					err,
				)
			}
		})

		t.Run("returns not found for wrong version", func(t *testing.T) {
			p := newTestProject(
				t,
				"40100000-0000-0000-0000-000000000001",
				"Remove Wrong Version",
				"started",
				[]iam.UserID{user1ID},
				2,
				time.Now().UTC(),
			)

			insertTestProjectDirectly(t, db, p)

			p = project.RestoreProject(
				p.ID(),
				p.Name(),
				p.Status(),
				p.Members(),
				1,
				p.CreatedAt(),
				p.UpdatedAt(),
			)

			err := repo.Remove(ctx, p)

			if !errors.Is(err, project.ErrProjectNotFound) {
				t.Fatalf(
					"expected ErrProjectNotFound, got %v",
					err,
				)
			}
		})
	})
}

func newTestProject(
	t *testing.T,
	id string,
	name string,
	status string,
	members []iam.UserID,
	version int,
	createdAt time.Time,
) *project.Project {
	t.Helper()

	projectID, err := project.NewProjectID(id)
	if err != nil {
		t.Fatal(err)
	}

	nameVO, err := project.NewName(name)
	if err != nil {
		t.Fatal(err)
	}

	statusVO, err := project.NewStatus(status)
	if err != nil {
		t.Fatal(err)
	}

	return project.RestoreProject(
		projectID,
		nameVO,
		statusVO,
		members,
		version,
		createdAt,
		createdAt,
	)
}

func newTestProjectUser(
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
		) VALUES ($1, $2, $3, $4, $5)`,
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

func insertTestProjectDirectly(
	t *testing.T,
	db *sql.DB,
	p *project.Project,
) {
	t.Helper()

	_, err := db.ExecContext(
		context.Background(),
		`INSERT INTO projects(
			id,
			status,
			name,
			version,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		p.ID().String(),
		p.Status().String(),
		p.Name().String(),
		p.Version(),
		p.CreatedAt(),
		p.UpdatedAt(),
	)
	if err != nil {
		t.Fatalf("failed to insert test project: %v", err)
	}

	for _, memberID := range p.Members() {
		_, err := db.ExecContext(
			context.Background(),
			`INSERT INTO project_members(
				project_id,
				user_id
			) VALUES ($1, $2)`,
			p.ID().String(),
			memberID.String(),
		)
		if err != nil {
			t.Fatalf("failed to insert test project member: %v", err)
		}
	}
}

func assertProjectEqual(
	t *testing.T,
	expected *project.Project,
	actual *project.Project,
) {
	t.Helper()

	if actual == nil {
		t.Fatal("expected project, got nil")
	}

	if actual.ID() != expected.ID() {
		t.Fatalf(
			"expected ID %s, got %s",
			expected.ID(),
			actual.ID(),
		)
	}

	if actual.Name() != expected.Name() {
		t.Fatalf(
			"expected name %s, got %s",
			expected.Name(),
			actual.Name(),
		)
	}

	if actual.Status() != expected.Status() {
		t.Fatalf(
			"expected status %s, got %s",
			expected.Status(),
			actual.Status(),
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

	assertProjectMembers(
		t,
		actual,
		expected.Members(),
	)
}

func assertProjectMembers(
	t *testing.T,
	p *project.Project,
	expected []iam.UserID,
) {
	t.Helper()

	actual := p.Members()

	if len(actual) != len(expected) {
		t.Fatalf(
			"expected %d members, got %d",
			len(expected),
			len(actual),
		)
	}

	expectedSet := make(map[iam.UserID]struct{}, len(expected))

	for _, id := range expected {
		expectedSet[id] = struct{}{}
	}

	for _, id := range actual {
		if _, ok := expectedSet[id]; !ok {
			t.Fatalf(
				"unexpected member %s",
				id,
			)
		}
	}
}
