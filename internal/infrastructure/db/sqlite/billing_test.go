package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"mizu/internal/domain/billing"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"

	"github.com/stretchr/testify/assert"
)

func TestBillingRepository(t *testing.T) {
	db := newTestSQLite(t)
	repo := NewBillingRepository(db)
	ctx := context.Background()

	t.Run("Add", func(t *testing.T) {
		t.Run("adds invoice with items", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000001",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			invoice := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000002",
				projectID,
				true,
				billing.StatusPending,
				"Add Invoice 505000000000000000000000000000000002",
				now,
			)

			if err := repo.Add(ctx, invoice); err != nil {
				t.Fatalf("failed to add invoice: %v", err)
			}

			got, err := repo.Get(ctx, invoice.ID())
			if err != nil {
				t.Fatalf("failed to get invoice: %v", err)
			}

			assertBillingInvoiceEqual(t, invoice, got)
		})

		t.Run("adds quote with items", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000011",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			invoice := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000012",
				projectID,
				false,
				billing.StatusPending,
				"Add Quote 505000000000000000000000000000000012",
				now,
			)

			if err := repo.Add(ctx, invoice); err != nil {
				t.Fatalf("failed to add quote: %v", err)
			}

			got, err := repo.Get(ctx, invoice.ID())
			if err != nil {
				t.Fatalf("failed to get quote: %v", err)
			}

			assertBillingInvoiceEqual(t, invoice, got)
		})

		t.Run("rejects duplicate invoice id", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000021",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			invoice := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000022",
				projectID,
				true,
				billing.StatusPending,
				"Duplicate Invoice 505000000000000000000000000000000022",
				now,
			)

			if err := repo.Add(ctx, invoice); err != nil {
				t.Fatalf("failed to add first invoice: %v", err)
			}

			err := repo.Add(ctx, invoice)
			if err == nil {
				t.Fatal("expected duplicate invoice error")
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("returns invoice with all fields and items", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000031",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			invoice := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000032",
				projectID,
				true,
				billing.StatusAccepted,
				"Get Invoice 505000000000000000000000000000000032",
				now,
			)

			if err := repo.Add(ctx, invoice); err != nil {
				t.Fatalf("failed to add invoice: %v", err)
			}

			got, err := repo.Get(ctx, invoice.ID())
			if err != nil {
				t.Fatalf("failed to get invoice: %v", err)
			}

			assertBillingInvoiceEqual(t, invoice, got)
		})

		t.Run("returns not found for missing invoice", func(t *testing.T) {
			invoiceID, err := billing.NewInvoiceID(
				"50500000-0000-0000-0000-000000000041",
			)
			if err != nil {
				t.Fatal(err)
			}

			_, err = repo.Get(ctx, invoiceID)
			if err == nil {
				t.Fatal("expected invoice not found error")
			}

			if !errors.Is(err, billing.ErrBillingInvoiceNotFound) {
				t.Fatalf(
					"expected %v, got %v",
					billing.ErrBillingInvoiceNotFound,
					err,
				)
			}
		})
	})

	t.Run("GetCurrencyByCode", func(t *testing.T) {
		t.Run("returns currency", func(t *testing.T) {
			code, err := billing.NewCurrencyCode("USD")
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.GetCurrencyByCode(ctx, code)
			if err != nil {
				t.Fatalf("failed to get currency: %v", err)
			}

			if got.Code().String() != "USD" {
				t.Fatalf("expected USD, got %s", got.Code())
			}

			if got.Name().String() != "United States Dollar" {
				t.Fatalf("expected United States Dollar, got %s", got.Name())
			}

			if got.Symbol().String() != "$" {
				t.Fatalf("expected $, got %s", got.Symbol())
			}

			if got.DecimalPlaces().Int() != 2 {
				t.Fatalf(
					"expected 2 decimal places, got %d",
					got.DecimalPlaces().Int(),
				)
			}
		})

		t.Run("returns not found for missing currency", func(t *testing.T) {
			code, err := billing.NewCurrencyCode("ZZZ")
			if err != nil {
				t.Fatal(err)
			}

			_, err = repo.GetCurrencyByCode(ctx, code)
			if err == nil {
				t.Fatal("expected currency not found error")
			}

			if !errors.Is(err, billing.ErrBillingCurrencyNotFound) {
				t.Fatalf(
					"expected %v, got %v",
					billing.ErrBillingCurrencyNotFound,
					err,
				)
			}
		})
	})

	t.Run("GetStatsByProject", func(t *testing.T) {
		t.Run("returns invoice and quote statistics", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000051",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			invoices := []*billing.Invoice{
				newTestBillingInvoice(
					t,
					"50500000-0000-0000-0000-000000000052",
					projectID,
					true,
					billing.StatusPending,
					"Stats Pending 505000000000000000000000000000000052",
					now,
				),
				newTestBillingInvoice(
					t,
					"50500000-0000-0000-0000-000000000053",
					projectID,
					true,
					billing.StatusPaid,
					"Stats Paid 505000000000000000000000000000000053",
					now.Add(time.Minute),
				),
				newTestBillingInvoice(
					t,
					"50500000-0000-0000-0000-000000000054",
					projectID,
					false,
					billing.StatusPending,
					"Stats Quote 505000000000000000000000000000000054",
					now.Add(2*time.Minute),
				),
			}

			for _, invoice := range invoices {
				if err := repo.Add(ctx, invoice); err != nil {
					t.Fatalf("failed to add invoice: %v", err)
				}
			}

			stats, err := repo.GetStatsByProject(ctx, projectID)
			if err != nil {
				t.Fatalf("failed to get project stats: %v", err)
			}

			if stats.InvoiceCount() != 2 {
				t.Fatalf(
					"expected 2 invoices, got %d",
					stats.InvoiceCount(),
				)
			}

			if stats.InvoicePaidCount() != 1 {
				t.Fatalf(
					"expected 1 paid invoice, got %d",
					stats.InvoicePaidCount(),
				)
			}

			if stats.QuoteCount() != 1 {
				t.Fatalf(
					"expected 1 quote, got %d",
					stats.QuoteCount(),
				)
			}
		})
	})

	t.Run("ListByProject", func(t *testing.T) {
		t.Run("lists invoices for project", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000061",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			first := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000062",
				projectID,
				true,
				billing.StatusPending,
				"Project Invoice First 505000000000000000000000000000000062",
				now,
			)

			second := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000063",
				projectID,
				false,
				billing.StatusRejected,
				"Project Quote Second 505000000000000000000000000000000063",
				now.Add(time.Minute),
			)

			for _, invoice := range []*billing.Invoice{first, second} {
				if err := repo.Add(ctx, invoice); err != nil {
					t.Fatalf("failed to add invoice: %v", err)
				}
			}

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.ListByProject(
				ctx,
				billing.FilterByProject{
					ProjectID: projectID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list invoices: %v", err)
			}

			if len(got) != 2 {
				t.Fatalf("expected 2 invoices, got %d", len(got))
			}

			if got[0].ID() != second.ID() {
				t.Fatalf(
					"expected newest invoice %s first, got %s",
					second.ID(),
					got[0].ID(),
				)
			}

			if got[1].ID() != first.ID() {
				t.Fatalf(
					"expected older invoice %s second, got %s",
					first.ID(),
					got[1].ID(),
				)
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000071",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			match := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000072",
				projectID,
				true,
				billing.StatusPending,
				"Keyword Match 505000000000000000000000000000000072",
				now,
			)

			other := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000073",
				projectID,
				true,
				billing.StatusPending,
				"Completely Different 505000000000000000000000000000000073",
				now.Add(time.Minute),
			)

			for _, invoice := range []*billing.Invoice{match, other} {
				if err := repo.Add(ctx, invoice); err != nil {
					t.Fatalf("failed to add invoice: %v", err)
				}
			}

			keyword := "Keyword Match"

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.ListByProject(
				ctx,
				billing.FilterByProject{
					ProjectID: projectID,
					Keyword:   &keyword,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list invoices: %v", err)
			}

			if len(got) != 1 {
				t.Fatalf("expected 1 invoice, got %d", len(got))
			}

			if got[0].ID() != match.ID() {
				t.Fatalf(
					"expected invoice %s, got %s",
					match.ID(),
					got[0].ID(),
				)
			}
		})

		t.Run("filters by invoice type", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000081",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			invoice := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000082",
				projectID,
				true,
				billing.StatusPending,
				"Invoice Filter 505000000000000000000000000000000082",
				now,
			)

			quote := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000083",
				projectID,
				false,
				billing.StatusPending,
				"Quote Filter 505000000000000000000000000000000083",
				now.Add(time.Minute),
			)

			for _, item := range []*billing.Invoice{invoice, quote} {
				if err := repo.Add(ctx, item); err != nil {
					t.Fatalf("failed to add invoice: %v", err)
				}
			}

			isInvoice := true

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.ListByProject(
				ctx,
				billing.FilterByProject{
					ProjectID: projectID,
					IsInvoice: &isInvoice,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list invoices: %v", err)
			}

			if len(got) != 1 {
				t.Fatalf("expected 1 invoice, got %d", len(got))
			}

			if got[0].ID() != invoice.ID() {
				t.Fatalf(
					"expected invoice %s, got %s",
					invoice.ID(),
					got[0].ID(),
				)
			}
		})

		t.Run("filters by status", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000091",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			pending := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000092",
				projectID,
				true,
				billing.StatusPending,
				"Status Pending 505000000000000000000000000000000092",
				now,
			)

			paid := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000093",
				projectID,
				true,
				billing.StatusPaid,
				"Status Paid 505000000000000000000000000000000093",
				now.Add(time.Minute),
			)

			for _, invoice := range []*billing.Invoice{pending, paid} {
				if err := repo.Add(ctx, invoice); err != nil {
					t.Fatalf("failed to add invoice: %v", err)
				}
			}

			status := billing.StatusPaid

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.ListByProject(
				ctx,
				billing.FilterByProject{
					ProjectID: projectID,
					Status:    &status,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list invoices: %v", err)
			}

			if len(got) != 1 {
				t.Fatalf("expected 1 invoice, got %d", len(got))
			}

			if got[0].ID() != paid.ID() {
				t.Fatalf(
					"expected invoice %s, got %s",
					paid.ID(),
					got[0].ID(),
				)
			}
		})

		t.Run("returns empty slice when no invoices exist", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000101",
			)

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.ListByProject(
				ctx,
				billing.FilterByProject{
					ProjectID: projectID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list invoices: %v", err)
			}

			if got == nil {
				t.Fatal("expected non-nil empty slice")
			}

			if len(got) != 0 {
				t.Fatalf("expected 0 invoices, got %d", len(got))
			}
		})
	})

	t.Run("ListByMember", func(t *testing.T) {
		t.Run("lists invoices from member projects", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000111",
			)

			memberID := newTestContractUser(
				t,
				db,
				"50500000-0000-0000-0000-000000000112",
			)

			insertTestBillingProjectMember(
				t,
				db,
				projectID,
				memberID,
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			invoice := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000113",
				projectID,
				true,
				billing.StatusPending,
				"Member Invoice 505000000000000000000000000000000113",
				now,
			)

			if err := repo.Add(ctx, invoice); err != nil {
				t.Fatalf("failed to add invoice: %v", err)
			}

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.ListByMember(
				ctx,
				billing.FilterByMember{
					MemberID: memberID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list member invoices: %v", err)
			}

			if len(got) != 1 {
				t.Fatalf("expected 1 invoice, got %d", len(got))
			}

			assertBillingInvoiceEqual(t, invoice, got[0])
		})

		t.Run("does not return invoices from unrelated projects", func(t *testing.T) {
			memberProjectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000121",
			)

			otherProjectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000122",
			)

			memberID := newTestContractUser(
				t,
				db,
				"50500000-0000-0000-0000-000000000123",
			)

			insertTestBillingProjectMember(
				t,
				db,
				memberProjectID,
				memberID,
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			memberInvoice := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000124",
				memberProjectID,
				true,
				billing.StatusPending,
				"Member Project Invoice 505000000000000000000000000000000124",
				now,
			)

			otherInvoice := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000125",
				otherProjectID,
				true,
				billing.StatusPending,
				"Other Project Invoice 505000000000000000000000000000000125",
				now.Add(time.Minute),
			)

			for _, invoice := range []*billing.Invoice{
				memberInvoice,
				otherInvoice,
			} {
				if err := repo.Add(ctx, invoice); err != nil {
					t.Fatalf("failed to add invoice: %v", err)
				}
			}

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.ListByMember(
				ctx,
				billing.FilterByMember{
					MemberID: memberID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list member invoices: %v", err)
			}

			if len(got) != 1 {
				t.Fatalf("expected 1 invoice, got %d", len(got))
			}

			if got[0].ID() != memberInvoice.ID() {
				t.Fatalf(
					"expected invoice %s, got %s",
					memberInvoice.ID(),
					got[0].ID(),
				)
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000131",
			)

			memberID := newTestContractUser(
				t,
				db,
				"50500000-0000-0000-0000-000000000132",
			)

			insertTestBillingProjectMember(
				t,
				db,
				projectID,
				memberID,
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			match := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000133",
				projectID,
				true,
				billing.StatusPending,
				"Member Keyword Match 505000000000000000000000000000000133",
				now,
			)

			other := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000134",
				projectID,
				true,
				billing.StatusPending,
				"Something Else 505000000000000000000000000000000134",
				now.Add(time.Minute),
			)

			for _, invoice := range []*billing.Invoice{match, other} {
				if err := repo.Add(ctx, invoice); err != nil {
					t.Fatalf("failed to add invoice: %v", err)
				}
			}

			keyword := "Member Keyword Match"

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.ListByMember(
				ctx,
				billing.FilterByMember{
					MemberID: memberID,
					Keyword:  &keyword,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list member invoices: %v", err)
			}

			if len(got) != 1 {
				t.Fatalf("expected 1 invoice, got %d", len(got))
			}

			if got[0].ID() != match.ID() {
				t.Fatalf(
					"expected invoice %s, got %s",
					match.ID(),
					got[0].ID(),
				)
			}
		})
	})

	t.Run("CountByProject", func(t *testing.T) {
		t.Run("counts invoices for project", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000141",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			for i, id := range []string{
				"50500000-0000-0000-0000-000000000142",
				"50500000-0000-0000-0000-000000000143",
				"50500000-0000-0000-0000-000000000144",
			} {
				invoice := newTestBillingInvoice(
					t,
					id,
					projectID,
					true,
					billing.StatusPending,
					"Count Project "+string(rune('A'+i)),
					now.Add(time.Duration(i)*time.Minute),
				)

				if err := repo.Add(ctx, invoice); err != nil {
					t.Fatalf("failed to add invoice: %v", err)
				}
			}

			count, err := repo.CountByProject(
				ctx,
				billing.FilterByProject{
					ProjectID: projectID,
				},
			)
			if err != nil {
				t.Fatalf("failed to count invoices: %v", err)
			}

			if count != 3 {
				t.Fatalf("expected 3 invoices, got %d", count)
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000151",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			match := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000152",
				projectID,
				true,
				billing.StatusPending,
				"Count Keyword Match 505000000000000000000000000000000152",
				now,
			)

			other := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000153",
				projectID,
				true,
				billing.StatusPending,
				"Count Other 505000000000000000000000000000000153",
				now.Add(time.Minute),
			)

			for _, invoice := range []*billing.Invoice{match, other} {
				if err := repo.Add(ctx, invoice); err != nil {
					t.Fatalf("failed to add invoice: %v", err)
				}
			}

			keyword := "Count Keyword Match"

			count, err := repo.CountByProject(
				ctx,
				billing.FilterByProject{
					ProjectID: projectID,
					Keyword:   &keyword,
				},
			)
			if err != nil {
				t.Fatalf("failed to count invoices: %v", err)
			}

			if count != 1 {
				t.Fatalf("expected 1 invoice, got %d", count)
			}
		})
	})

	t.Run("CountByMember", func(t *testing.T) {
		t.Run("counts invoices for member projects", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000161",
			)

			memberID := newTestContractUser(
				t,
				db,
				"50500000-0000-0000-0000-000000000162",
			)

			insertTestBillingProjectMember(
				t,
				db,
				projectID,
				memberID,
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			for _, id := range []string{
				"50500000-0000-0000-0000-000000000163",
				"50500000-0000-0000-0000-000000000164",
			} {
				invoice := newTestBillingInvoice(
					t,
					id,
					projectID,
					true,
					billing.StatusPending,
					"Member Count Invoice "+id,
					now,
				)

				if err := repo.Add(ctx, invoice); err != nil {
					t.Fatalf("failed to add invoice: %v", err)
				}
			}

			count, err := repo.CountByMember(
				ctx,
				billing.FilterByMember{
					MemberID: memberID,
				},
			)
			if err != nil {
				t.Fatalf("failed to count member invoices: %v", err)
			}

			if count != 2 {
				t.Fatalf("expected 2 invoices, got %d", count)
			}
		})
	})

	t.Run("Save", func(t *testing.T) {
		t.Run("updates invoice fields and items", func(t *testing.T) {
			project1ID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000171",
			)

			project2ID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000172",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			invoice := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000173",
				project1ID,
				true,
				billing.StatusPending,
				"Save Original 505000000000000000000000000000000173",
				now,
			)

			if err := repo.Add(ctx, invoice); err != nil {
				t.Fatalf("failed to add invoice: %v", err)
			}

			updatedAt := now.Add(time.Minute)

			updated := newTestBillingInvoiceWithItems(
				t,
				invoice.ID().String(),
				project2ID,
				true,
				billing.StatusPaid,
				"Save Updated 505000000000000000000000000000000173",
				updatedAt,
				[]billing.Item{
					newTestBillingItem(
						t,
						"Updated Development",
						"3",
						"150",
						"5",
						billing.DiscountTypeFixed,
						"10",
						billing.TaxTypeFixed,
					),
				},
			)

			restored := billing.RestoreInvoice(
				updated.ID(),
				updated.ProjectID(),
				updated.IsInvoice(),
				updated.Status(),
				updated.DueAt(),
				updated.Currency(),
				updated.Note(),
				updated.Items(),
				updated.TotalTax(),
				updated.TotalDiscount(),
				updated.SubTotal(),
				invoice.Version(),
				invoice.CreatedAt(),
				updatedAt,
			)

			updated = &restored

			if err := repo.Save(ctx, updated); err != nil {
				t.Fatalf("failed to save invoice: %v", err)
			}

			got, err := repo.Get(ctx, invoice.ID())
			if err != nil {
				t.Fatalf("failed to get saved invoice: %v", err)
			}

			if got.ProjectID() != project2ID {
				t.Fatalf(
					"expected project %s, got %s",
					project2ID,
					got.ProjectID(),
				)
			}

			if got.IsInvoice() != updated.IsInvoice() {
				t.Fatalf(
					"expected is_invoice %v, got %v",
					updated.IsInvoice(),
					got.IsInvoice(),
				)
			}

			if got.Status() != updated.Status() {
				t.Fatalf(
					"expected status %s, got %s",
					updated.Status(),
					got.Status(),
				)
			}

			if got.Note() == nil || updated.Note() == nil {
				t.Fatal("expected notes to be non-nil")
			}

			if *got.Note() != *updated.Note() {
				t.Fatalf(
					"expected note %q, got %q",
					*updated.Note(),
					*got.Note(),
				)
			}

			if got.Version() != invoice.Version()+1 {
				t.Fatalf(
					"expected version %d, got %d",
					invoice.Version()+1,
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

			if len(got.Items()) != 1 {
				t.Fatalf("expected 1 item, got %d", len(got.Items()))
			}

			assertBillingItemEqual(
				t,
				updated.Items()[0],
				got.Items()[0],
			)
		})

		t.Run("replaces existing items", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000181",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			invoice := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000182",
				projectID,
				true,
				billing.StatusPending,
				"Replace Items 505000000000000000000000000000000182",
				now,
			)

			if err := repo.Add(ctx, invoice); err != nil {
				t.Fatalf("failed to add invoice: %v", err)
			}

			updated := billing.RestoreInvoice(
				invoice.ID(),
				invoice.ProjectID(),
				invoice.IsInvoice(),
				billing.StatusPaid,
				invoice.DueAt(),
				invoice.Currency(),
				invoice.Note(),
				[]billing.Item{
					newTestBillingItem(
						t,
						"Replacement Item",
						"4",
						"200",
						"10",
						billing.DiscountTypeFixed,
						"5",
						billing.TaxTypeFixed,
					),
				},
				invoice.TotalTax(),
				invoice.TotalDiscount(),
				invoice.SubTotal(),
				invoice.Version(),
				invoice.CreatedAt(),
				now.Add(time.Minute),
			)

			if err := repo.Save(ctx, &updated); err != nil {
				t.Fatalf("failed to save invoice: %v", err)
			}

			got, err := repo.Get(ctx, invoice.ID())
			if err != nil {
				t.Fatalf("failed to get invoice: %v", err)
			}

			if len(got.Items()) != 1 {
				t.Fatalf("expected 1 item, got %d", len(got.Items()))
			}

			assertBillingItemEqual(
				t,
				updated.Items()[0],
				got.Items()[0],
			)
		})

		t.Run("returns concurrent modification error", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000191",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			invoice := newTestBillingInvoice(
				t,
				"50500000-0000-0000-0000-000000000192",
				projectID,
				true,
				billing.StatusPending,
				"Concurrent 505000000000000000000000000000000192",
				now,
			)

			if err := repo.Add(ctx, invoice); err != nil {
				t.Fatalf("failed to add invoice: %v", err)
			}

			stale := billing.RestoreInvoice(
				invoice.ID(),
				invoice.ProjectID(),
				invoice.IsInvoice(),
				billing.StatusPaid,
				invoice.DueAt(),
				invoice.Currency(),
				invoice.Note(),
				invoice.Items(),
				invoice.TotalTax(),
				invoice.TotalDiscount(),
				invoice.SubTotal(),
				invoice.Version()-1,
				invoice.CreatedAt(),
				now.Add(time.Minute),
			)

			err := repo.Save(ctx, &stale)
			if err == nil {
				t.Fatal("expected concurrent modification error")
			}

			if !errors.Is(err, billing.ErrBillingConcurrentModification) {
				t.Fatalf(
					"expected %v, got %v",
					billing.ErrBillingConcurrentModification,
					err,
				)
			}
		})
	})

	t.Run("ListStatsByProjects", func(t *testing.T) {
		t.Run("returns statistics for projects", func(t *testing.T) {
			project1ID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000201",
			)

			project2ID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000202",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			fixtures := []*billing.Invoice{
				newTestBillingInvoice(
					t,
					"50500000-0000-0000-0000-000000000203",
					project1ID,
					true,
					billing.StatusPending,
					"Project Stats 1A 505000000000000000000000000000000203",
					now,
				),
				newTestBillingInvoice(
					t,
					"50500000-0000-0000-0000-000000000204",
					project1ID,
					true,
					billing.StatusPaid,
					"Project Stats 1B 505000000000000000000000000000000204",
					now.Add(time.Minute),
				),
				newTestBillingInvoice(
					t,
					"50500000-0000-0000-0000-000000000205",
					project1ID,
					false,
					billing.StatusPending,
					"Project Stats 1C 505000000000000000000000000000000205",
					now.Add(2*time.Minute),
				),
				newTestBillingInvoice(
					t,
					"50500000-0000-0000-0000-000000000206",
					project2ID,
					true,
					billing.StatusPaid,
					"Project Stats 2A 505000000000000000000000000000000206",
					now.Add(3*time.Minute),
				),
			}

			for _, invoice := range fixtures {
				if err := repo.Add(ctx, invoice); err != nil {
					t.Fatalf("failed to add invoice: %v", err)
				}
			}

			stats, err := repo.ListStatsByProjects(
				ctx,
				[]project.ProjectID{
					project1ID,
					project2ID,
				},
			)
			if err != nil {
				t.Fatalf("failed to list project stats: %v", err)
			}

			project1Stats, ok := stats[project1ID]
			if !ok {
				t.Fatal("expected statistics for project 1")
			}

			if project1Stats.InvoiceCount() != 3 {
				t.Fatalf(
					"expected project 1 to have 3 invoices, got %d",
					project1Stats.InvoiceCount(),
				)
			}

			if project1Stats.InvoicePaidCount() != 1 {
				t.Fatalf(
					"expected project 1 to have 1 paid invoice, got %d",
					project1Stats.InvoicePaidCount(),
				)
			}

			if project1Stats.QuoteCount() != 1 {
				t.Fatalf(
					"expected project 1 to have 1 quote, got %d",
					project1Stats.QuoteCount(),
				)
			}

			project2Stats, ok := stats[project2ID]
			if !ok {
				t.Fatal("expected statistics for project 2")
			}

			if project2Stats.InvoiceCount() != 1 {
				t.Fatalf(
					"expected project 2 to have 1 invoice, got %d",
					project2Stats.InvoiceCount(),
				)
			}

			if project2Stats.InvoicePaidCount() != 1 {
				t.Fatalf(
					"expected project 2 to have 1 paid invoice, got %d",
					project2Stats.InvoicePaidCount(),
				)
			}
		})

		t.Run("returns empty map for no projects", func(t *testing.T) {
			stats, err := repo.ListStatsByProjects(ctx, nil)
			if err != nil {
				t.Fatalf("failed to list project stats: %v", err)
			}

			if stats == nil {
				t.Fatal("expected non-nil map")
			}

			if len(stats) != 0 {
				t.Fatalf("expected empty map, got %d entries", len(stats))
			}
		})
	})

	t.Run("ListMonthlyPaidCountByProject", func(t *testing.T) {
		t.Run("returns twelve months", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50500000-0000-0000-0000-000000000211",
			)

			metrics, err := repo.ListMonthlyPaidCountByProject(
				ctx,
				projectID,
			)
			if err != nil {
				t.Fatalf("failed to list monthly paid count: %v", err)
			}

			if len(metrics) != 12 {
				t.Fatalf(
					"expected 12 metrics, got %d",
					len(metrics),
				)
			}
		})
	})

	t.Run("ListMonthlyPaidCountByMember", func(t *testing.T) {
		t.Run("returns twelve months", func(t *testing.T) {
			memberID := newTestContractUser(
				t,
				db,
				"50500000-0000-0000-0000-000000000221",
			)

			metrics, err := repo.ListMonthlyPaidCountByMember(
				ctx,
				memberID,
			)
			if err != nil {
				t.Fatalf("failed to list monthly paid count: %v", err)
			}

			if len(metrics) != 12 {
				t.Fatalf(
					"expected 12 metrics, got %d",
					len(metrics),
				)
			}
		})
	})
}

func newTestBillingInvoice(
	t *testing.T,
	id string,
	projectID project.ProjectID,
	isInvoice bool,
	status billing.Status,
	note string,
	now time.Time,
) *billing.Invoice {
	t.Helper()

	return newTestBillingInvoiceWithItems(
		t,
		id,
		projectID,
		isInvoice,
		status,
		note,
		now,
		[]billing.Item{
			newTestBillingItem(
				t,
				"Development",
				"2",
				"100",
				"10",
				billing.DiscountTypeFixed,
				"5",
				billing.TaxTypeFixed,
			),
			newTestBillingItem(
				t,
				"Consultation",
				"1",
				"100",
				"10",
				billing.DiscountTypeFixed,
				"5",
				billing.TaxTypeFixed,
			),
		},
	)
}

func newTestBillingInvoiceWithItems(
	t *testing.T,
	id string,
	projectID project.ProjectID,
	isInvoice bool,
	status billing.Status,
	note string,
	now time.Time,
	items []billing.Item,
) *billing.Invoice {
	t.Helper()

	invoiceID, err := billing.NewInvoiceID(id)
	if err != nil {
		t.Fatal(err)
	}

	currencyName, err := billing.NewCurrencyName("United States Dollar")
	if err != nil {
		t.Fatal(err)
	}

	currencySymbol, err := billing.NewCurrencySymbol("$")
	if err != nil {
		t.Fatal(err)
	}

	currencyCode, err := billing.NewCurrencyCode("USD")
	if err != nil {
		t.Fatal(err)
	}

	currencyDecimals, err := billing.NewCurrencyDecimals(2)
	if err != nil {
		t.Fatal(err)
	}

	currency := billing.NewCurrency(
		currencyName,
		currencySymbol,
		currencyCode,
		currencyDecimals,
	)

	noteValue := note

	totalTax, err := billing.NewDecimal("20")
	if err != nil {
		t.Fatal(err)
	}

	totalDiscount, err := billing.NewDecimal("40")
	if err != nil {
		t.Fatal(err)
	}

	subTotal, err := billing.NewDecimal("280")
	if err != nil {
		t.Fatal(err)
	}

	returnPtr := billing.RestoreInvoice(
		invoiceID,
		projectID,
		isInvoice,
		status,
		nil,
		currency,
		&noteValue,
		items,
		totalTax,
		totalDiscount,
		subTotal,
		1,
		now,
		now,
	)

	return &returnPtr
}

func newTestBillingItem(
	t *testing.T,
	description string,
	qty string,
	unitPrice string,
	discountRate string,
	discountType billing.DiscountType,
	taxRate string,
	taxType billing.TaxType,
) billing.Item {
	t.Helper()

	q, err := billing.NewQty(qty)
	if err != nil {
		t.Fatal(err)
	}

	price, err := billing.NewDecimal(unitPrice)
	if err != nil {
		t.Fatal(err)
	}

	discount, err := billing.NewDecimal(discountRate)
	if err != nil {
		t.Fatal(err)
	}

	tax, err := billing.NewDecimal(taxRate)
	if err != nil {
		t.Fatal(err)
	}

	zero, err := billing.NewDecimal("0")
	if err != nil {
		t.Fatal(err)
	}

	gross, err := billing.NewDecimal("200")
	if err != nil {
		t.Fatal(err)
	}

	lineDiscount, err := billing.NewDecimal("20")
	if err != nil {
		t.Fatal(err)
	}

	lineNet, err := billing.NewDecimal("180")
	if err != nil {
		t.Fatal(err)
	}

	lineTax, err := billing.NewDecimal("10")
	if err != nil {
		t.Fatal(err)
	}

	lineTotal, err := billing.NewDecimal("190")
	if err != nil {
		t.Fatal(err)
	}

	return billing.RestoreItem(
		description,
		q,
		price,
		discount,
		discountType,
		tax,
		taxType,
		zero,
		price,
		zero,
		lineDiscount,
		lineTax,
		gross,
		lineNet,
		lineTotal,
	)
}

func insertTestBillingProjectMember(
	t *testing.T,
	db *sql.DB,
	projectID project.ProjectID,
	userID iam.UserID,
) {
	t.Helper()

	_, err := db.Exec(
		`
		INSERT INTO project_members (
			project_id,
			user_id
		) VALUES ($1, $2)
		`,
		projectID.String(),
		userID.String(),
	)
	if err != nil {
		t.Fatalf("failed to insert project member: %v", err)
	}
}

func assertBillingInvoiceEqual(t *testing.T, want, got *billing.Invoice) {
	t.Helper()

	assert.Equal(t, want.ID(), got.ID())
	assert.Equal(t, want.ProjectID(), got.ProjectID())
	assert.Equal(t, want.IsInvoice(), got.IsInvoice())
	assert.Equal(t, want.Status(), got.Status())
	assert.Equal(t, want.Currency(), got.Currency())
	assert.Equal(t, want.Note(), got.Note())
	assert.Equal(t, want.Version(), got.Version())
	assert.Equal(t, want.CreatedAt(), got.CreatedAt())
	assert.Equal(t, want.UpdatedAt(), got.UpdatedAt())

	assert.True(t, want.TotalTax().Equals(got.TotalTax()))
	assert.True(t, want.TotalDiscount().Equals(got.TotalDiscount()))
	assert.True(t, want.SubTotal().Equals(got.SubTotal()))

	wantItems := want.Items()
	gotItems := got.Items()

	assert.Len(t, gotItems, len(wantItems))

	for i := range wantItems {
		assertBillingItemEqual(t, wantItems[i], gotItems[i])
	}
}

func assertBillingItemEqual(t *testing.T, want billing.Item, got billing.Item) {
	t.Helper()

	assert.Equal(t, want.Description(), got.Description())
	assert.True(t, want.Qty().ToDecimal().Equals(got.Qty().ToDecimal()))
	assert.True(t, want.UnitPrice().Equals(got.UnitPrice()))
	assert.True(t, want.DiscountRate().Equals(got.DiscountRate()))
	assert.Equal(t, want.DiscountType(), got.DiscountType())
	assert.True(t, want.TaxRate().Equals(got.TaxRate()))
	assert.Equal(t, want.TaxType(), got.TaxType())
	assert.True(t, want.DiscountAmountPerUnit().Equals(got.DiscountAmountPerUnit()))
	assert.True(t, want.TaxableBasePerUnit().Equals(got.TaxableBasePerUnit()))
	assert.True(t, want.TaxAmountPerUnit().Equals(got.TaxAmountPerUnit()))
	assert.True(t, want.LineGross().Equals(got.LineGross()))
	assert.True(t, want.LineDiscount().Equals(got.LineDiscount()))
	assert.True(t, want.LineNet().Equals(got.LineNet()))
	assert.True(t, want.LineTax().Equals(got.LineTax()))
	assert.True(t, want.LineTotal().Equals(got.LineTotal()))
}
