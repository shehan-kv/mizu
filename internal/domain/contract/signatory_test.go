package contract

import (
	"testing"
	"time"

	"mizu/internal/domain/iam"
)

func TestNewSignatory(t *testing.T) {
	now := time.Now()
	userID := iam.UserID("user-1")

	signatory := NewSignatory(userID, now)

	if signatory.UserID() != userID {
		t.Fatalf("expected user ID %q, got %q", userID, signatory.UserID())
	}

	if signatory.Status() != SignatoryStatusPending {
		t.Fatalf(
			"expected status %q, got %q",
			SignatoryStatusPending,
			signatory.Status(),
		)
	}

	if !signatory.UpdatedAt().Equal(now) {
		t.Fatalf("expected UpdatedAt %v, got %v", now, signatory.UpdatedAt())
	}
}

func TestRestoreSignatory(t *testing.T) {
	now := time.Now()
	userID := iam.UserID("user-1")

	signatory := RestoreSignatory(
		userID,
		SignatoryStatusSigned,
		now,
	)

	if signatory.UserID() != userID {
		t.Fatalf("expected user ID %q, got %q", userID, signatory.UserID())
	}

	if signatory.Status() != SignatoryStatusSigned {
		t.Fatalf(
			"expected status %q, got %q",
			SignatoryStatusSigned,
			signatory.Status(),
		)
	}

	if !signatory.UpdatedAt().Equal(now) {
		t.Fatalf("expected UpdatedAt %v, got %v", now, signatory.UpdatedAt())
	}
}
