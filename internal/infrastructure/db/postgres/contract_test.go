package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"mizu/internal/domain/common"
	"mizu/internal/domain/contract"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

func TestContractRepository(t *testing.T) {
	db := newTestPostgres(t)
	repo := NewContractRepository(db)
	ctx := context.Background()

	t.Run("Add", func(t *testing.T) {
		t.Run("adds contract with signatories", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40000000-0000-0000-0000-000000000001",
			)

			user1ID := newTestContractUser(
				t,
				db,
				"40000000-0000-0000-0000-000000000002",
			)

			user2ID := newTestContractUser(
				t,
				db,
				"40000000-0000-0000-0000-000000000003",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c := newTestContract(
				t,
				"40000000-0000-0000-0000-000000000004",
				projectID,
				"Test Contract",
				"pending",
				1,
				[]iam.UserID{user1ID, user2ID},
				now,
			)

			if err := repo.Add(ctx, c); err != nil {
				t.Fatalf("failed to add contract: %v", err)
			}

			got, err := repo.Get(ctx, c.ID())
			if err != nil {
				t.Fatalf("failed to get inserted contract: %v", err)
			}

			assertContractEqual(t, c, got)
		})

		t.Run("returns error for duplicate contract", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40000000-0000-0000-0000-000000000021",
			)

			userID := newTestContractUser(
				t,
				db,
				"40000000-0000-0000-0000-000000000023",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			signatories := []iam.UserID{userID}

			first := newTestContract(
				t,
				"40000000-0000-0000-0000-000000000022",
				projectID,
				"Duplicate Contract",
				"pending",
				1,
				signatories,
				now,
			)

			if err := repo.Add(ctx, first); err != nil {
				t.Fatalf("failed to add first contract: %v", err)
			}

			second := newTestContract(
				t,
				"40000000-0000-0000-0000-000000000022",
				projectID,
				"Duplicate Contract",
				"pending",
				1,
				signatories,
				now.Add(time.Second),
			)

			err := repo.Add(ctx, second)
			if err == nil {
				t.Fatal("expected duplicate contract error, got nil")
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("returns contract with signatories", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40100000-0000-0000-0000-000000000001",
			)

			user1ID := newTestContractUser(
				t,
				db,
				"40100000-0000-0000-0000-000000000002",
			)

			user2ID := newTestContractUser(
				t,
				db,
				"40100000-0000-0000-0000-000000000003",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c := newTestContract(
				t,
				"40100000-0000-0000-0000-000000000004",
				projectID,
				"Get Contract",
				"pending",
				3,
				[]iam.UserID{user1ID, user2ID},
				now,
			)

			insertTestContractDirectly(t, db, c)

			got, err := repo.Get(ctx, c.ID())
			if err != nil {
				t.Fatalf("failed to get contract: %v", err)
			}

			assertContractEqual(t, c, got)
		})

		t.Run("returns contract without signatories", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40100000-0000-0000-0000-000000000011",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c := newTestContract(
				t,
				"40100000-0000-0000-0000-000000000012",
				projectID,
				"Empty Contract",
				"pending",
				1,
				nil,
				now,
			)

			insertTestContractDirectly(t, db, c)

			got, err := repo.Get(ctx, c.ID())
			if err != nil {
				t.Fatalf("failed to get contract: %v", err)
			}

			assertContractEqual(t, c, got)
		})

		t.Run("returns not found", func(t *testing.T) {
			id, err := contract.NewContractID(
				"40100000-0000-0000-0000-000000000021",
			)
			if err != nil {
				t.Fatal(err)
			}

			_, err = repo.Get(ctx, id)
			if err == nil {
				t.Fatal("expected not found error, got nil")
			}

			if !errors.Is(err, contract.ErrContractNotFound) {
				t.Fatalf(
					"expected %v, got %v",
					contract.ErrContractNotFound,
					err,
				)
			}
		})
	})

	t.Run("GetStatsByProject", func(t *testing.T) {
		t.Run("returns total and signed counts", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40200000-0000-0000-0000-000000000001",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestContract(
				t,
				"40200000-0000-0000-0000-000000000002",
				projectID,
				"Pending One",
				"pending",
				1,
				nil,
				now,
			)

			c2 := newTestContract(
				t,
				"40200000-0000-0000-0000-000000000003",
				projectID,
				"Signed One",
				contract.StatusSigned.String(),
				1,
				nil,
				now.Add(time.Second),
			)

			c3 := newTestContract(
				t,
				"40200000-0000-0000-0000-000000000004",
				projectID,
				"Signed Two",
				contract.StatusSigned.String(),
				1,
				nil,
				now.Add(2*time.Second),
			)

			insertTestContractDirectly(t, db, c1)
			insertTestContractDirectly(t, db, c2)
			insertTestContractDirectly(t, db, c3)

			stats, err := repo.GetStatsByProject(ctx, projectID)
			if err != nil {
				t.Fatalf("failed to get stats: %v", err)
			}

			if stats.Total() != 3 {
				t.Fatalf("expected total 3, got %d", stats.Total())
			}

			if stats.Signed() != 2 {
				t.Fatalf("expected signed 2, got %d", stats.Signed())
			}
		})

		t.Run("returns zero stats when project has no contracts", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40200000-0000-0000-0000-000000000011",
			)

			stats, err := repo.GetStatsByProject(ctx, projectID)
			if err != nil {
				t.Fatalf("failed to get stats: %v", err)
			}

			if stats.Total() != 0 {
				t.Fatalf("expected total 0, got %d", stats.Total())
			}

			if stats.Signed() != 0 {
				t.Fatalf("expected signed 0, got %d", stats.Signed())
			}
		})
	})

	t.Run("ListByProject", func(t *testing.T) {
		t.Run("lists contracts for project with signatories", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40300000-0000-0000-0000-000000000001",
			)

			user1ID := newTestContractUser(
				t,
				db,
				"40300000-0000-0000-0000-000000000002",
			)

			user2ID := newTestContractUser(
				t,
				db,
				"40300000-0000-0000-0000-000000000003",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestContract(
				t,
				"40300000-0000-0000-0000-000000000004",
				projectID,
				"Project Contract One",
				"pending",
				1,
				[]iam.UserID{user1ID},
				now,
			)

			c2 := newTestContract(
				t,
				"40300000-0000-0000-0000-000000000005",
				projectID,
				"Project Contract Two",
				contract.StatusSigned.String(),
				2,
				[]iam.UserID{user1ID, user2ID},
				now.Add(time.Second),
			)

			otherProjectID := newTestContractProject(
				t,
				db,
				"40300000-0000-0000-0000-000000000006",
			)

			c3 := newTestContract(
				t,
				"40300000-0000-0000-0000-000000000007",
				otherProjectID,
				"Other Project Contract",
				"pending",
				1,
				nil,
				now.Add(2*time.Second),
			)

			insertTestContractDirectly(t, db, c1)
			insertTestContractDirectly(t, db, c2)
			insertTestContractDirectly(t, db, c3)

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			contracts, err := repo.ListByProject(
				ctx,
				contract.FilterByProject{
					ProjectID: projectID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list contracts: %v", err)
			}

			if len(contracts) != 2 {
				t.Fatalf("expected 2 contracts, got %d", len(contracts))
			}

			if contracts[0].ID() != c2.ID() {
				t.Fatalf(
					"expected first contract %s, got %s",
					c2.ID(),
					contracts[0].ID(),
				)
			}

			if contracts[1].ID() != c1.ID() {
				t.Fatalf(
					"expected second contract %s, got %s",
					c1.ID(),
					contracts[1].ID(),
				)
			}

			assertContractEqual(t, c2, contracts[0])
			assertContractEqual(t, c1, contracts[1])
		})

		t.Run("filters by keyword", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40300000-0000-0000-0000-000000000011",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestContract(
				t,
				"40300000-0000-0000-0000-000000000012",
				projectID,
				"Website Contract",
				"pending",
				1,
				nil,
				now,
			)

			c2 := newTestContract(
				t,
				"40300000-0000-0000-0000-000000000013",
				projectID,
				"Mobile Contract",
				"pending",
				1,
				nil,
				now.Add(time.Second),
			)

			insertTestContractDirectly(t, db, c1)
			insertTestContractDirectly(t, db, c2)

			keyword := "Website"

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			contracts, err := repo.ListByProject(
				ctx,
				contract.FilterByProject{
					ProjectID: projectID,
					Keyword:   &keyword,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list contracts: %v", err)
			}

			if len(contracts) != 1 {
				t.Fatalf("expected 1 contract, got %d", len(contracts))
			}

			if contracts[0].ID() != c1.ID() {
				t.Fatalf(
					"expected %s, got %s",
					c1.ID(),
					contracts[0].ID(),
				)
			}
		})

		t.Run("filters by status", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40300000-0000-0000-0000-000000000021",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			pending := newTestContract(
				t,
				"40300000-0000-0000-0000-000000000022",
				projectID,
				"Pending Contract",
				"pending",
				1,
				nil,
				now,
			)

			signed := newTestContract(
				t,
				"40300000-0000-0000-0000-000000000023",
				projectID,
				"Signed Contract",
				contract.StatusSigned.String(),
				1,
				nil,
				now.Add(time.Second),
			)

			insertTestContractDirectly(t, db, pending)
			insertTestContractDirectly(t, db, signed)

			status, err := contract.NewStatus(
				contract.StatusSigned.String(),
			)
			if err != nil {
				t.Fatal(err)
			}

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			contracts, err := repo.ListByProject(
				ctx,
				contract.FilterByProject{
					ProjectID: projectID,
					Status:    &status,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list contracts: %v", err)
			}

			if len(contracts) != 1 {
				t.Fatalf("expected 1 contract, got %d", len(contracts))
			}

			if contracts[0].ID() != signed.ID() {
				t.Fatalf(
					"expected %s, got %s",
					signed.ID(),
					contracts[0].ID(),
				)
			}
		})

		t.Run("applies pagination", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40300000-0000-0000-0000-000000000031",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestContract(
				t,
				"40300000-0000-0000-0000-000000000032",
				projectID,
				"Pagination One",
				"pending",
				1,
				nil,
				now,
			)

			c2 := newTestContract(
				t,
				"40300000-0000-0000-0000-000000000033",
				projectID,
				"Pagination Two",
				"pending",
				1,
				nil,
				now.Add(time.Second),
			)

			c3 := newTestContract(
				t,
				"40300000-0000-0000-0000-000000000034",
				projectID,
				"Pagination Three",
				"pending",
				1,
				nil,
				now.Add(2*time.Second),
			)

			insertTestContractDirectly(t, db, c1)
			insertTestContractDirectly(t, db, c2)
			insertTestContractDirectly(t, db, c3)

			page, err := common.NewPage(1, 1)
			if err != nil {
				t.Fatal(err)
			}

			contracts, err := repo.ListByProject(
				ctx,
				contract.FilterByProject{
					ProjectID: projectID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list contracts: %v", err)
			}

			if len(contracts) != 1 {
				t.Fatalf("expected 1 contract, got %d", len(contracts))
			}

			if contracts[0].ID() != c2.ID() {
				t.Fatalf(
					"expected %s, got %s",
					c2.ID(),
					contracts[0].ID(),
				)
			}
		})

		t.Run("returns empty slice when project has no contracts", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40300000-0000-0000-0000-000000000041",
			)

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			contracts, err := repo.ListByProject(
				ctx,
				contract.FilterByProject{
					ProjectID: projectID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list contracts: %v", err)
			}

			if contracts == nil {
				t.Fatal("expected empty slice, got nil")
			}

			if len(contracts) != 0 {
				t.Fatalf("expected 0 contracts, got %d", len(contracts))
			}
		})
	})

	t.Run("ListBySignatory", func(t *testing.T) {
		t.Run("lists contracts for signatory", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40400000-0000-0000-0000-000000000001",
			)

			signatoryID := newTestContractUser(
				t,
				db,
				"40400000-0000-0000-0000-000000000002",
			)

			otherSignatoryID := newTestContractUser(
				t,
				db,
				"40400000-0000-0000-0000-000000000003",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestContract(
				t,
				"40400000-0000-0000-0000-000000000004",
				projectID,
				"Signatory One",
				"pending",
				1,
				[]iam.UserID{signatoryID},
				now,
			)

			c2 := newTestContract(
				t,
				"40400000-0000-0000-0000-000000000005",
				projectID,
				"Signatory Two",
				contract.StatusSigned.String(),
				1,
				[]iam.UserID{signatoryID, otherSignatoryID},
				now.Add(time.Second),
			)

			c3 := newTestContract(
				t,
				"40400000-0000-0000-0000-000000000006",
				projectID,
				"Other Signatory",
				"pending",
				1,
				[]iam.UserID{otherSignatoryID},
				now.Add(2*time.Second),
			)

			insertTestContractDirectly(t, db, c1)
			insertTestContractDirectly(t, db, c2)
			insertTestContractDirectly(t, db, c3)

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			contracts, err := repo.ListBySignatory(
				ctx,
				contract.FilterBySignatory{
					SignatoryID: signatoryID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list contracts: %v", err)
			}

			if len(contracts) != 2 {
				t.Fatalf("expected 2 contracts, got %d", len(contracts))
			}

			if contracts[0].ID() != c2.ID() {
				t.Fatalf(
					"expected first contract %s, got %s",
					c2.ID(),
					contracts[0].ID(),
				)
			}

			if contracts[1].ID() != c1.ID() {
				t.Fatalf(
					"expected second contract %s, got %s",
					c1.ID(),
					contracts[1].ID(),
				)
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40400000-0000-0000-0000-000000000011",
			)

			signatoryID := newTestContractUser(
				t,
				db,
				"40400000-0000-0000-0000-000000000012",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestContract(
				t,
				"40400000-0000-0000-0000-000000000013",
				projectID,
				"Website Agreement",
				"pending",
				1,
				[]iam.UserID{signatoryID},
				now,
			)

			c2 := newTestContract(
				t,
				"40400000-0000-0000-0000-000000000014",
				projectID,
				"Mobile Agreement",
				"pending",
				1,
				[]iam.UserID{signatoryID},
				now.Add(time.Second),
			)

			insertTestContractDirectly(t, db, c1)
			insertTestContractDirectly(t, db, c2)

			keyword := "Website"

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			contracts, err := repo.ListBySignatory(
				ctx,
				contract.FilterBySignatory{
					SignatoryID: signatoryID,
					Keyword:     &keyword,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list contracts: %v", err)
			}

			if len(contracts) != 1 {
				t.Fatalf("expected 1 contract, got %d", len(contracts))
			}

			if contracts[0].ID() != c1.ID() {
				t.Fatalf(
					"expected %s, got %s",
					c1.ID(),
					contracts[0].ID(),
				)
			}
		})

		t.Run("filters by status", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40400000-0000-0000-0000-000000000021",
			)

			signatoryID := newTestContractUser(
				t,
				db,
				"40400000-0000-0000-0000-000000000022",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			pending := newTestContract(
				t,
				"40400000-0000-0000-0000-000000000023",
				projectID,
				"Pending Agreement",
				"pending",
				1,
				[]iam.UserID{signatoryID},
				now,
			)

			signed := newTestContract(
				t,
				"40400000-0000-0000-0000-000000000024",
				projectID,
				"Signed Agreement",
				contract.StatusSigned.String(),
				1,
				[]iam.UserID{signatoryID},
				now.Add(time.Second),
			)

			insertTestContractDirectly(t, db, pending)
			insertTestContractDirectly(t, db, signed)

			status, err := contract.NewStatus(
				contract.StatusSigned.String(),
			)
			if err != nil {
				t.Fatal(err)
			}

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			contracts, err := repo.ListBySignatory(
				ctx,
				contract.FilterBySignatory{
					SignatoryID: signatoryID,
					Status:      &status,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list contracts: %v", err)
			}

			if len(contracts) != 1 {
				t.Fatalf("expected 1 contract, got %d", len(contracts))
			}

			if contracts[0].ID() != signed.ID() {
				t.Fatalf(
					"expected %s, got %s",
					signed.ID(),
					contracts[0].ID(),
				)
			}
		})

		t.Run("applies pagination", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40400000-0000-0000-0000-000000000031",
			)

			signatoryID := newTestContractUser(
				t,
				db,
				"40400000-0000-0000-0000-000000000032",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestContract(
				t,
				"40400000-0000-0000-0000-000000000033",
				projectID,
				"Pagination Test 404000000000000000000000000000000000033 One",
				"pending",
				1,
				[]iam.UserID{signatoryID},
				now,
			)

			c2 := newTestContract(
				t,
				"40400000-0000-0000-0000-000000000034",
				projectID,
				"Pagination Test 404000000000000000000000000000000000034 Two",
				"pending",
				1,
				[]iam.UserID{signatoryID},
				now.Add(time.Second),
			)

			c3 := newTestContract(
				t,
				"40400000-0000-0000-0000-000000000035",
				projectID,
				"Pagination Test 404000000000000000000000000000000000035 Three",
				"pending",
				1,
				[]iam.UserID{signatoryID},
				now.Add(2*time.Second),
			)

			insertTestContractDirectly(t, db, c1)
			insertTestContractDirectly(t, db, c2)
			insertTestContractDirectly(t, db, c3)

			page, err := common.NewPage(1, 1)
			if err != nil {
				t.Fatal(err)
			}

			contracts, err := repo.ListBySignatory(
				ctx,
				contract.FilterBySignatory{
					SignatoryID: signatoryID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list contracts: %v", err)
			}

			if len(contracts) != 1 {
				t.Fatalf("expected 1 contract, got %d", len(contracts))
			}

			if contracts[0].ID() != c2.ID() {
				t.Fatalf(
					"expected %s, got %s",
					c2.ID(),
					contracts[0].ID(),
				)
			}
		})

		t.Run("returns empty slice when signatory has no contracts", func(t *testing.T) {
			signatoryID := newTestContractUser(
				t,
				db,
				"40400000-0000-0000-0000-000000000041",
			)

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			contracts, err := repo.ListBySignatory(
				ctx,
				contract.FilterBySignatory{
					SignatoryID: signatoryID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list contracts: %v", err)
			}

			if contracts == nil {
				t.Fatal("expected empty slice, got nil")
			}

			if len(contracts) != 0 {
				t.Fatalf("expected 0 contracts, got %d", len(contracts))
			}
		})
	})

	t.Run("ListStatsByProjects", func(t *testing.T) {
		t.Run("returns stats grouped by project", func(t *testing.T) {
			project1ID := newTestContractProject(
				t,
				db,
				"40500000-0000-0000-0000-000000000001",
			)

			project2ID := newTestContractProject(
				t,
				db,
				"40500000-0000-0000-0000-000000000002",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestContract(
				t,
				"40500000-0000-0000-0000-000000000003",
				project1ID,
				"Project One Pending",
				"pending",
				1,
				nil,
				now,
			)

			c2 := newTestContract(
				t,
				"40500000-0000-0000-0000-000000000004",
				project1ID,
				"Project One Signed",
				contract.StatusSigned.String(),
				1,
				nil,
				now.Add(time.Second),
			)

			c3 := newTestContract(
				t,
				"40500000-0000-0000-0000-000000000005",
				project2ID,
				"Project Two Signed",
				contract.StatusSigned.String(),
				1,
				nil,
				now.Add(2*time.Second),
			)

			insertTestContractDirectly(t, db, c1)
			insertTestContractDirectly(t, db, c2)
			insertTestContractDirectly(t, db, c3)

			stats, err := repo.ListStatsByProjects(
				ctx,
				[]project.ProjectID{
					project1ID,
					project2ID,
				},
			)
			if err != nil {
				t.Fatalf("failed to list stats: %v", err)
			}

			if len(stats) != 2 {
				t.Fatalf("expected 2 projects, got %d", len(stats))
			}

			project1Stats, ok := stats[project1ID]
			if !ok {
				t.Fatalf("missing stats for project %s", project1ID)
			}

			if project1Stats.Total() != 2 {
				t.Fatalf(
					"expected project 1 total 2, got %d",
					project1Stats.Total(),
				)
			}

			if project1Stats.Signed() != 1 {
				t.Fatalf(
					"expected project 1 signed 1, got %d",
					project1Stats.Signed(),
				)
			}

			project2Stats, ok := stats[project2ID]
			if !ok {
				t.Fatalf("missing stats for project %s", project2ID)
			}

			if project2Stats.Total() != 1 {
				t.Fatalf(
					"expected project 2 total 1, got %d",
					project2Stats.Total(),
				)
			}

			if project2Stats.Signed() != 1 {
				t.Fatalf(
					"expected project 2 signed 1, got %d",
					project2Stats.Signed(),
				)
			}
		})

		t.Run("returns empty map for empty project IDs", func(t *testing.T) {
			stats, err := repo.ListStatsByProjects(ctx, nil)
			if err != nil {
				t.Fatalf("failed to list stats: %v", err)
			}

			if stats == nil {
				t.Fatal("expected empty map, got nil")
			}

			if len(stats) != 0 {
				t.Fatalf("expected empty map, got %d entries", len(stats))
			}
		})
	})

	t.Run("CountByProject", func(t *testing.T) {
		t.Run("counts contracts for project", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40600000-0000-0000-0000-000000000001",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestContract(
				t,
				"40600000-0000-0000-0000-000000000002",
				projectID,
				"Count One",
				"pending",
				1,
				nil,
				now,
			)

			c2 := newTestContract(
				t,
				"40600000-0000-0000-0000-000000000003",
				projectID,
				"Count Two",
				contract.StatusSigned.String(),
				1,
				nil,
				now.Add(time.Second),
			)

			otherProjectID := newTestContractProject(
				t,
				db,
				"40600000-0000-0000-0000-000000000004",
			)

			c3 := newTestContract(
				t,
				"40600000-0000-0000-0000-000000000005",
				otherProjectID,
				"Other Count",
				"pending",
				1,
				nil,
				now.Add(2*time.Second),
			)

			insertTestContractDirectly(t, db, c1)
			insertTestContractDirectly(t, db, c2)
			insertTestContractDirectly(t, db, c3)

			count, err := repo.CountByProject(
				ctx,
				contract.FilterByProject{
					ProjectID: projectID,
				},
			)
			if err != nil {
				t.Fatalf("failed to count contracts: %v", err)
			}

			if count != 2 {
				t.Fatalf("expected 2, got %d", count)
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40600000-0000-0000-0000-000000000011",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestContract(
				t,
				"40600000-0000-0000-0000-000000000012",
				projectID,
				"Website Contract 406000000000000000000000000000000000012",
				"pending",
				1,
				nil,
				now,
			)

			c2 := newTestContract(
				t,
				"40600000-0000-0000-0000-000000000013",
				projectID,
				"Mobile Contract 406000000000000000000000000000000000013",
				"pending",
				1,
				nil,
				now.Add(time.Second),
			)

			insertTestContractDirectly(t, db, c1)
			insertTestContractDirectly(t, db, c2)

			keyword := "Website"

			count, err := repo.CountByProject(
				ctx,
				contract.FilterByProject{
					ProjectID: projectID,
					Keyword:   &keyword,
				},
			)
			if err != nil {
				t.Fatalf("failed to count contracts: %v", err)
			}

			if count != 1 {
				t.Fatalf("expected 1, got %d", count)
			}
		})

		t.Run("filters by status", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40600000-0000-0000-0000-000000000021",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			pending := newTestContract(
				t,
				"40600000-0000-0000-0000-000000000022",
				projectID,
				"Pending Contract 40600000000000000000000000000022",
				"pending",
				1,
				nil,
				now,
			)

			signed := newTestContract(
				t,
				"40600000-0000-0000-0000-000000000023",
				projectID,
				"Signed Contract 40600000000000000000000000000023",
				contract.StatusSigned.String(),
				1,
				nil,
				now.Add(time.Second),
			)

			insertTestContractDirectly(t, db, pending)
			insertTestContractDirectly(t, db, signed)

			status, err := contract.NewStatus(
				contract.StatusSigned.String(),
			)
			if err != nil {
				t.Fatal(err)
			}

			count, err := repo.CountByProject(
				ctx,
				contract.FilterByProject{
					ProjectID: projectID,
					Status:    &status,
				},
			)
			if err != nil {
				t.Fatalf("failed to count contracts: %v", err)
			}

			if count != 1 {
				t.Fatalf("expected 1, got %d", count)
			}
		})
	})

	t.Run("CountBySignatory", func(t *testing.T) {
		t.Run("counts contracts for signatory", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40700000-0000-0000-0000-000000000001",
			)

			signatoryID := newTestContractUser(
				t,
				db,
				"40700000-0000-0000-0000-000000000002",
			)

			otherSignatoryID := newTestContractUser(
				t,
				db,
				"40700000-0000-0000-0000-000000000003",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestContract(
				t,
				"40700000-0000-0000-0000-000000000004",
				projectID,
				"Count One 40700000000000000000000000000004",
				"pending",
				1,
				[]iam.UserID{signatoryID},
				now,
			)

			c2 := newTestContract(
				t,
				"40700000-0000-0000-0000-000000000005",
				projectID,
				"Count Two 40700000000000000000000000000005",
				contract.StatusSigned.String(),
				1,
				[]iam.UserID{signatoryID},
				now.Add(time.Second),
			)

			c3 := newTestContract(
				t,
				"40700000-0000-0000-0000-000000000006",
				projectID,
				"Other Count 40700000000000000000000000000006",
				"pending",
				1,
				[]iam.UserID{otherSignatoryID},
				now.Add(2*time.Second),
			)

			insertTestContractDirectly(t, db, c1)
			insertTestContractDirectly(t, db, c2)
			insertTestContractDirectly(t, db, c3)

			count, err := repo.CountBySignatory(
				ctx,
				contract.FilterBySignatory{
					SignatoryID: signatoryID,
				},
			)
			if err != nil {
				t.Fatalf("failed to count contracts: %v", err)
			}

			if count != 2 {
				t.Fatalf("expected 2, got %d", count)
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40700000-0000-0000-0000-000000000011",
			)

			signatoryID := newTestContractUser(
				t,
				db,
				"40700000-0000-0000-0000-000000000012",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestContract(
				t,
				"40700000-0000-0000-0000-000000000013",
				projectID,
				"Website Agreement 40700000000000000000000000000013",
				"pending",
				1,
				[]iam.UserID{signatoryID},
				now,
			)

			c2 := newTestContract(
				t,
				"40700000-0000-0000-0000-000000000014",
				projectID,
				"Mobile Agreement 40700000000000000000000000000014",
				"pending",
				1,
				[]iam.UserID{signatoryID},
				now.Add(time.Second),
			)

			insertTestContractDirectly(t, db, c1)
			insertTestContractDirectly(t, db, c2)

			keyword := "Website"

			count, err := repo.CountBySignatory(
				ctx,
				contract.FilterBySignatory{
					SignatoryID: signatoryID,
					Keyword:     &keyword,
				},
			)
			if err != nil {
				t.Fatalf("failed to count contracts: %v", err)
			}

			if count != 1 {
				t.Fatalf("expected 1, got %d", count)
			}
		})

		t.Run("filters by status", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40700000-0000-0000-0000-000000000021",
			)

			signatoryID := newTestContractUser(
				t,
				db,
				"40700000-0000-0000-0000-000000000022",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			pending := newTestContract(
				t,
				"40700000-0000-0000-0000-000000000023",
				projectID,
				"Pending Agreement 40700000000000000000000000000023",
				"pending",
				1,
				[]iam.UserID{signatoryID},
				now,
			)

			signed := newTestContract(
				t,
				"40700000-0000-0000-0000-000000000024",
				projectID,
				"Signed Agreement 40700000000000000000000000000024",
				contract.StatusSigned.String(),
				1,
				[]iam.UserID{signatoryID},
				now.Add(time.Second),
			)

			insertTestContractDirectly(t, db, pending)
			insertTestContractDirectly(t, db, signed)

			status, err := contract.NewStatus(
				contract.StatusSigned.String(),
			)
			if err != nil {
				t.Fatal(err)
			}

			count, err := repo.CountBySignatory(
				ctx,
				contract.FilterBySignatory{
					SignatoryID: signatoryID,
					Status:      &status,
				},
			)
			if err != nil {
				t.Fatalf("failed to count contracts: %v", err)
			}

			if count != 1 {
				t.Fatalf("expected 1, got %d", count)
			}
		})
	})

	t.Run("Save", func(t *testing.T) {
		t.Run("updates contract and replaces signatories", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40800000-0000-0000-0000-000000000001",
			)

			user1ID := newTestContractUser(
				t,
				db,
				"40800000-0000-0000-0000-000000000002",
			)

			user2ID := newTestContractUser(
				t,
				db,
				"40800000-0000-0000-0000-000000000003",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c := newTestContract(
				t,
				"40800000-0000-0000-0000-000000000004",
				projectID,
				"Original Contract",
				"pending",
				1,
				[]iam.UserID{user1ID},
				now,
			)

			insertTestContractDirectly(t, db, c)

			updatedAt := now.Add(time.Minute)

			updatedName, err := contract.NewName("Updated Contract")
			if err != nil {
				t.Fatal(err)
			}

			updatedStatus, err := contract.NewStatus(
				contract.StatusSigned.String(),
			)
			if err != nil {
				t.Fatal(err)
			}

			updatedTerms, err := contract.NewTerms("These are the standard terms and conditions for this contract.")
			if err != nil {
				t.Fatal(err)
			}

			updatedSignatory := contract.NewSignatory(
				user2ID,
				updatedAt,
			)

			updated := contract.RestoreContract(
				c.ID(),
				c.ProjectID(),
				updatedName,
				updatedStatus,
				updatedTerms,
				[]contract.Signatory{updatedSignatory},
				c.Version(),
				c.CreatedAt(),
				updatedAt,
			)

			if err := repo.Save(ctx, updated); err != nil {
				t.Fatalf("failed to save contract: %v", err)
			}

			got, err := repo.Get(ctx, c.ID())
			if err != nil {
				t.Fatalf("failed to get saved contract: %v", err)
			}

			if got.Name().String() != "Updated Contract" {
				t.Fatalf(
					"expected updated name, got %s",
					got.Name().String(),
				)
			}

			if got.Status().String() != contract.StatusSigned.String() {
				t.Fatalf(
					"expected status %s, got %s",
					contract.StatusSigned.String(),
					got.Status().String(),
				)
			}

			if got.Terms().String() != "These are the standard terms and conditions for this contract." {
				t.Fatalf(
					"expected updated terms, got %s",
					got.Terms().String(),
				)
			}

			if got.Version() != c.Version()+1 {
				t.Fatalf(
					"expected version %d, got %d",
					c.Version()+1,
					got.Version(),
				)
			}

			assertContractSignatories(
				t,
				got,
				[]iam.UserID{user2ID},
			)
		})

		t.Run("returns concurrent modification error", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"40800000-0000-0000-0000-000000000021",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c := newTestContract(
				t,
				"40800000-0000-0000-0000-000000000022",
				projectID,
				"Concurrent Contract",
				"pending",
				1,
				nil,
				now,
			)

			insertTestContractDirectly(t, db, c)

			stale := contract.RestoreContract(
				c.ID(),
				c.ProjectID(),
				c.Name(),
				c.Status(),
				c.Terms(),
				nil,
				c.Version()-1,
				c.CreatedAt(),
				now.Add(time.Minute),
			)

			err := repo.Save(ctx, stale)
			if err == nil {
				t.Fatal("expected concurrent modification error, got nil")
			}

			if !errors.Is(
				err,
				contract.ErrContractConcurrentModification,
			) {
				t.Fatalf(
					"expected %v, got %v",
					contract.ErrContractConcurrentModification,
					err,
				)
			}
		})
	})
}

func newTestContractProject(
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
			status,
			name,
			version,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		id,
		"started",
		"Contract Test Project "+id,
		1,
		now,
		now,
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

func newTestContractUser(
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
		"Contract",
		"Test User",
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

func newTestContract(
	t *testing.T,
	id string,
	projectID project.ProjectID,
	name string,
	status string,
	version int,
	userIDs []iam.UserID,
	createdAt time.Time,
) *contract.Contract {
	t.Helper()

	contractID, err := contract.NewContractID(id)
	if err != nil {
		t.Fatal(err)
	}

	nameVO, err := contract.NewName(name)
	if err != nil {
		t.Fatal(err)
	}

	statusVO, err := contract.NewStatus(status)
	if err != nil {
		t.Fatal(err)
	}

	termsVO, err := contract.NewTerms("These are the standard terms and conditions for this contract.")
	if err != nil {
		t.Fatal(err)
	}

	signatories := make([]contract.Signatory, 0, len(userIDs))

	for _, userID := range userIDs {
		signatories = append(
			signatories,
			contract.NewSignatory(userID, createdAt),
		)
	}

	return contract.RestoreContract(
		contractID,
		projectID,
		nameVO,
		statusVO,
		termsVO,
		signatories,
		version,
		createdAt,
		createdAt,
	)
}

func insertTestContractDirectly(
	t *testing.T,
	db *sql.DB,
	c *contract.Contract,
) {
	t.Helper()

	_, err := db.ExecContext(
		context.Background(),
		`INSERT INTO contracts(
			id,
			project_id,
			name,
			status,
			terms,
			version,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID().String(),
		c.ProjectID().String(),
		c.Name().String(),
		c.Status().String(),
		c.Terms().String(),
		c.Version(),
		c.CreatedAt(),
		c.UpdatedAt(),
	)
	if err != nil {
		t.Fatalf("failed to insert test contract: %v", err)
	}

	for _, signatory := range c.Signatories() {
		_, err := db.ExecContext(
			context.Background(),
			`INSERT INTO contract_signatories(
				contract_id,
				user_id,
				status,
				updated_at
			) VALUES ($1, $2, $3, $4)`,
			c.ID().String(),
			signatory.UserID().String(),
			signatory.Status().String(),
			signatory.UpdatedAt(),
		)
		if err != nil {
			t.Fatalf("failed to insert test signatory: %v", err)
		}
	}
}

func assertContractEqual(
	t *testing.T,
	expected *contract.Contract,
	actual *contract.Contract,
) {
	t.Helper()

	if actual == nil {
		t.Fatal("expected contract, got nil")
	}

	if actual.ID() != expected.ID() {
		t.Fatalf(
			"expected ID %s, got %s",
			expected.ID(),
			actual.ID(),
		)
	}

	if actual.ProjectID() != expected.ProjectID() {
		t.Fatalf(
			"expected project ID %s, got %s",
			expected.ProjectID(),
			actual.ProjectID(),
		)
	}

	if actual.Name().String() != expected.Name().String() {
		t.Fatalf(
			"expected name %s, got %s",
			expected.Name().String(),
			actual.Name().String(),
		)
	}

	if actual.Status().String() != expected.Status().String() {
		t.Fatalf(
			"expected status %s, got %s",
			expected.Status().String(),
			actual.Status().String(),
		)
	}

	if actual.Terms().String() != expected.Terms().String() {
		t.Fatalf(
			"expected terms %s, got %s",
			expected.Terms().String(),
			actual.Terms().String(),
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

	assertContractSignatories(
		t,
		actual,
		signatoryUserIDs(expected),
	)
}

func assertContractSignatories(
	t *testing.T,
	c *contract.Contract,
	expected []iam.UserID,
) {
	t.Helper()

	actual := c.Signatories()

	if len(actual) != len(expected) {
		t.Fatalf(
			"expected %d signatories, got %d",
			len(expected),
			len(actual),
		)
	}

	expectedSet := make(map[iam.UserID]struct{}, len(expected))

	for _, id := range expected {
		expectedSet[id] = struct{}{}
	}

	actualSet := make(map[iam.UserID]struct{}, len(actual))

	for _, signatory := range actual {
		actualSet[signatory.UserID()] = struct{}{}
	}

	if len(actualSet) != len(expectedSet) {
		t.Fatalf(
			"expected signatories %v, got %v",
			expected,
			signatoryUserIDs(c),
		)
	}

	for id := range expectedSet {
		if _, ok := actualSet[id]; !ok {
			t.Fatalf(
				"expected signatory %s, got %v",
				id,
				signatoryUserIDs(c),
			)
		}
	}
}

func signatoryUserIDs(c *contract.Contract) []iam.UserID {
	signatories := c.Signatories()

	ids := make([]iam.UserID, 0, len(signatories))

	for _, signatory := range signatories {
		ids = append(ids, signatory.UserID())
	}

	return ids
}
