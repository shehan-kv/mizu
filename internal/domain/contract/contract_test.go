package contract

import (
	"testing"
	"time"

	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

func newTestContract(t *testing.T, now time.Time) *Contract {
	t.Helper()

	contractID, err := NewContractID("contract-1")
	if err != nil {
		t.Fatalf("failed to create contract ID: %v", err)
	}

	name, err := NewName("Test Contract")
	if err != nil {
		t.Fatalf("failed to create contract name: %v", err)
	}

	terms, err := NewTerms(
		"These are valid contract terms that contain more than fifty characters.",
	)
	if err != nil {
		t.Fatalf("failed to create contract terms: %v", err)
	}

	projectID := project.ProjectID("project-1")

	signatories := []Signatory{
		NewSignatory(iam.UserID("user-1"), now),
		NewSignatory(iam.UserID("user-2"), now),
	}

	contract, err := NewContract(
		contractID,
		projectID,
		name,
		terms,
		signatories,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create contract: %v", err)
	}

	return contract
}

func TestNewContract(t *testing.T) {
	now := time.Now()

	t.Run("creates pending contract", func(t *testing.T) {
		contract := newTestContract(t, now)

		if contract.Status() != StatusPending {
			t.Fatalf(
				"expected status %q, got %q",
				StatusPending,
				contract.Status(),
			)
		}

		if contract.Version() != 1 {
			t.Fatalf("expected version 1, got %d", contract.Version())
		}

		if !contract.CreatedAt().Equal(now) {
			t.Fatalf("expected CreatedAt %v, got %v", now, contract.CreatedAt())
		}

		if !contract.UpdatedAt().Equal(now) {
			t.Fatalf("expected UpdatedAt %v, got %v", now, contract.UpdatedAt())
		}
	})

	t.Run("requires at least two signatories", func(t *testing.T) {
		id := ContractID("contract-1")
		name := Name{value: "Test Contract"}
		terms := Terms{
			value: "These are valid contract terms that contain more than fifty characters.",
		}

		signatories := []Signatory{
			NewSignatory(iam.UserID("user-1"), now),
		}

		_, err := NewContract(
			id,
			project.ProjectID("project-1"),
			name,
			terms,
			signatories,
			now,
		)

		if err != ErrContractMustHaveAtLeastTwoSignatories {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractMustHaveAtLeastTwoSignatories,
				err,
			)
		}
	})

	t.Run("rejects duplicate signatories", func(t *testing.T) {
		id := ContractID("contract-1")
		name := Name{value: "Test Contract"}
		terms := Terms{
			value: "These are valid contract terms that contain more than fifty characters.",
		}

		userID := iam.UserID("user-1")

		signatories := []Signatory{
			NewSignatory(userID, now),
			NewSignatory(userID, now),
		}

		_, err := NewContract(
			id,
			project.ProjectID("project-1"),
			name,
			terms,
			signatories,
			now,
		)

		if err != ErrContractSignatoriesMustBeUnique {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractSignatoriesMustBeUnique,
				err,
			)
		}
	})
}

func TestContractPullEvents(t *testing.T) {
	now := time.Now()
	contract := newTestContract(t, now)

	events := contract.PullEvents()

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event, ok := events[0].(ContractCreatedEvent)
	if !ok {
		t.Fatalf("expected ContractCreatedEvent, got %T", events[0])
	}

	if event.ContractID != contract.ID() {
		t.Fatalf("expected contract ID %q, got %q", contract.ID(), event.ContractID)
	}

	if event.ProjectID != contract.ProjectID() {
		t.Fatalf(
			"expected project ID %q, got %q",
			contract.ProjectID(),
			event.ProjectID,
		)
	}

	if event.EventType() != EventTypeContractCreated {
		t.Fatalf(
			"expected event type %q, got %q",
			EventTypeContractCreated,
			event.EventType(),
		)
	}

	if !event.OccurredAt.Equal(now) {
		t.Fatalf("expected OccurredAt %v, got %v", now, event.OccurredAt)
	}

	events = contract.PullEvents()

	if len(events) != 0 {
		t.Fatalf("expected events to be cleared, got %d events", len(events))
	}
}

func TestContractSign(t *testing.T) {
	t.Run("signs pending signatory", func(t *testing.T) {
		now := time.Now()
		contract := newTestContract(t, now)

		signTime := now.Add(time.Second)

		err := contract.Sign(iam.UserID("user-1"), signTime)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if contract.Status() != StatusPending {
			t.Fatalf(
				"expected contract to remain pending, got %q",
				contract.Status(),
			)
		}

		signatories := contract.Signatories()

		if signatories[0].Status() != SignatoryStatusSigned {
			t.Fatalf(
				"expected signatory to be signed, got %q",
				signatories[0].Status(),
			)
		}

		if !signatories[0].UpdatedAt().Equal(signTime) {
			t.Fatalf(
				"expected signatory UpdatedAt %v, got %v",
				signTime,
				signatories[0].UpdatedAt(),
			)
		}

		if !contract.UpdatedAt().Equal(signTime) {
			t.Fatalf(
				"expected contract UpdatedAt %v, got %v",
				signTime,
				contract.UpdatedAt(),
			)
		}
	})

	t.Run("moves contract to signed when all signatories sign", func(t *testing.T) {
		now := time.Now()
		contract := newTestContract(t, now)

		firstSignTime := now.Add(time.Second)
		secondSignTime := now.Add(2 * time.Second)

		if err := contract.Sign(iam.UserID("user-1"), firstSignTime); err != nil {
			t.Fatalf("first sign failed: %v", err)
		}

		if err := contract.Sign(iam.UserID("user-2"), secondSignTime); err != nil {
			t.Fatalf("second sign failed: %v", err)
		}

		if contract.Status() != StatusSigned {
			t.Fatalf(
				"expected status %q, got %q",
				StatusSigned,
				contract.Status(),
			)
		}

		if !contract.UpdatedAt().Equal(secondSignTime) {
			t.Fatalf(
				"expected UpdatedAt %v, got %v",
				secondSignTime,
				contract.UpdatedAt(),
			)
		}
	})

	t.Run("rejects signing non-pending contract", func(t *testing.T) {
		now := time.Now()
		contract := newTestContract(t, now)

		signTime := now.Add(time.Second)

		if err := contract.Reject(iam.UserID("user-1"), signTime); err != nil {
			t.Fatalf("reject failed: %v", err)
		}

		err := contract.Sign(iam.UserID("user-2"), now.Add(2*time.Second))
		if err != ErrContractSignRequiresPending {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractSignRequiresPending,
				err,
			)
		}
	})

	t.Run("rejects unknown signatory", func(t *testing.T) {
		now := time.Now()
		contract := newTestContract(t, now)

		err := contract.Sign(iam.UserID("unknown"), now.Add(time.Second))
		if err != ErrContractSignatoryNotFound {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractSignatoryNotFound,
				err,
			)
		}

		if contract.Status() != StatusPending {
			t.Fatalf("contract should remain pending")
		}
	})

	t.Run("rejects signatory that already acted", func(t *testing.T) {
		now := time.Now()
		contract := newTestContract(t, now)

		signTime := now.Add(time.Second)

		if err := contract.Sign(iam.UserID("user-1"), signTime); err != nil {
			t.Fatalf("first sign failed: %v", err)
		}

		err := contract.Sign(iam.UserID("user-1"), now.Add(2*time.Second))
		if err != ErrContractSignatoryAlreadyActed {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractSignatoryAlreadyActed,
				err,
			)
		}
	})
}

func TestContractReject(t *testing.T) {
	t.Run("rejects contract", func(t *testing.T) {
		now := time.Now()
		contract := newTestContract(t, now)

		rejectTime := now.Add(time.Second)

		err := contract.Reject(iam.UserID("user-1"), rejectTime)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if contract.Status() != StatusRejected {
			t.Fatalf(
				"expected status %q, got %q",
				StatusRejected,
				contract.Status(),
			)
		}

		signatories := contract.Signatories()

		if signatories[0].Status() != SignatoryStatusRejected {
			t.Fatalf(
				"expected signatory status %q, got %q",
				SignatoryStatusRejected,
				signatories[0].Status(),
			)
		}

		if !signatories[0].UpdatedAt().Equal(rejectTime) {
			t.Fatalf(
				"expected signatory UpdatedAt %v, got %v",
				rejectTime,
				signatories[0].UpdatedAt(),
			)
		}

		if !contract.UpdatedAt().Equal(rejectTime) {
			t.Fatalf(
				"expected contract UpdatedAt %v, got %v",
				rejectTime,
				contract.UpdatedAt(),
			)
		}
	})

	t.Run("rejects rejecting non-pending contract", func(t *testing.T) {
		now := time.Now()
		contract := newTestContract(t, now)

		rejectTime := now.Add(time.Second)

		if err := contract.Reject(iam.UserID("user-1"), rejectTime); err != nil {
			t.Fatalf("reject failed: %v", err)
		}

		err := contract.Reject(iam.UserID("user-2"), now.Add(2*time.Second))
		if err != ErrContractRejectRequiresPending {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractRejectRequiresPending,
				err,
			)
		}
	})

	t.Run("rejects unknown signatory", func(t *testing.T) {
		now := time.Now()
		contract := newTestContract(t, now)

		err := contract.Reject(iam.UserID("unknown"), now.Add(time.Second))
		if err != ErrContractSignatoryNotFound {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractSignatoryNotFound,
				err,
			)
		}
	})
}

func TestContractReplaceSignatories(t *testing.T) {
	t.Run("replaces pending signatories", func(t *testing.T) {
		now := time.Now()
		contract := newTestContract(t, now)

		replaceTime := now.Add(time.Second)

		err := contract.ReplaceSignatories(
			[]iam.UserID{
				iam.UserID("user-3"),
				iam.UserID("user-4"),
			},
			replaceTime,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		signatories := contract.Signatories()

		if len(signatories) != 2 {
			t.Fatalf("expected 2 signatories, got %d", len(signatories))
		}

		if signatories[0].UserID() != iam.UserID("user-3") {
			t.Fatalf("unexpected first signatory")
		}

		if signatories[1].UserID() != iam.UserID("user-4") {
			t.Fatalf("unexpected second signatory")
		}

		if signatories[0].Status() != SignatoryStatusPending ||
			signatories[1].Status() != SignatoryStatusPending {
			t.Fatalf("replacement signatories should be pending")
		}

		if !contract.UpdatedAt().Equal(replaceTime) {
			t.Fatalf(
				"expected UpdatedAt %v, got %v",
				replaceTime,
				contract.UpdatedAt(),
			)
		}
	})

	t.Run("requires at least two signatories", func(t *testing.T) {
		now := time.Now()
		contract := newTestContract(t, now)

		err := contract.ReplaceSignatories(
			[]iam.UserID{iam.UserID("user-3")},
			now.Add(time.Second),
		)

		if err != ErrContractMustHaveAtLeastTwoSignatories {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractMustHaveAtLeastTwoSignatories,
				err,
			)
		}
	})

	t.Run("rejects duplicate user IDs", func(t *testing.T) {
		now := time.Now()
		contract := newTestContract(t, now)

		userID := iam.UserID("user-3")

		err := contract.ReplaceSignatories(
			[]iam.UserID{userID, userID},
			now.Add(time.Second),
		)

		if err != ErrContractSignatoriesMustBeUnique {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractSignatoriesMustBeUnique,
				err,
			)
		}
	})

	t.Run("cannot replace signatories after one has acted", func(t *testing.T) {
		now := time.Now()
		contract := newTestContract(t, now)

		if err := contract.Sign(iam.UserID("user-1"), now.Add(time.Second)); err != nil {
			t.Fatalf("sign failed: %v", err)
		}

		err := contract.ReplaceSignatories(
			[]iam.UserID{
				iam.UserID("user-3"),
				iam.UserID("user-4"),
			},
			now.Add(2*time.Second),
		)

		if err != ErrContractSignatoriesLocked {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractSignatoriesLocked,
				err,
			)
		}
	})

	t.Run("cannot replace signatories after contract is rejected", func(t *testing.T) {
		now := time.Now()
		contract := newTestContract(t, now)

		if err := contract.Reject(iam.UserID("user-1"), now.Add(time.Second)); err != nil {
			t.Fatalf("reject failed: %v", err)
		}

		err := contract.ReplaceSignatories(
			[]iam.UserID{
				iam.UserID("user-3"),
				iam.UserID("user-4"),
			},
			now.Add(2*time.Second),
		)

		if err != ErrContractReplaceSignatoriesRequiresPending {
			t.Fatalf(
				"expected %v, got %v",
				ErrContractReplaceSignatoriesRequiresPending,
				err,
			)
		}
	})
}

func TestContractHasSignatory(t *testing.T) {
	now := time.Now()
	contract := newTestContract(t, now)

	if !contract.HasSignatory(iam.UserID("user-1")) {
		t.Fatal("expected user-1 to be a signatory")
	}

	if !contract.HasSignatory(iam.UserID("user-2")) {
		t.Fatal("expected user-2 to be a signatory")
	}

	if contract.HasSignatory(iam.UserID("user-3")) {
		t.Fatal("expected user-3 not to be a signatory")
	}
}

func TestContractSignatoriesReturnsCopy(t *testing.T) {
	now := time.Now()
	contract := newTestContract(t, now)

	signatories := contract.Signatories()

	signatories[0] = NewSignatory(
		iam.UserID("different-user"),
		now,
	)

	original := contract.Signatories()

	if original[0].UserID() != iam.UserID("user-1") {
		t.Fatal("modifying returned signatories should not modify the aggregate")
	}
}

func TestContractEquals(t *testing.T) {
	now := time.Now()

	first := newTestContract(t, now)

	second := RestoreContract(
		first.ID(),
		first.ProjectID(),
		first.Name(),
		first.Status(),
		first.Terms(),
		first.Signatories(),
		5,
		first.CreatedAt(),
		first.UpdatedAt(),
	)

	if !first.Equals(second) {
		t.Fatal("contracts with the same ID should be equal")
	}

	otherID := ContractID("contract-2")

	third := RestoreContract(
		otherID,
		first.ProjectID(),
		first.Name(),
		first.Status(),
		first.Terms(),
		first.Signatories(),
		first.Version(),
		first.CreatedAt(),
		first.UpdatedAt(),
	)

	if first.Equals(third) {
		t.Fatal("contracts with different IDs should not be equal")
	}
}
