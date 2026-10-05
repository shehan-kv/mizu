package eventhandler

import (
	"context"
	"errors"
	domainiam "mizu/internal/domain/iam"
	"testing"
	"time"
)

func TestEmailChangedHandle(t *testing.T) {
	t.Run("ignores unrelated event", func(t *testing.T) {
		m := &fakeMailer{}
		idGen := &fakeIDGenerator{}
		vRepo := &fakeVerificationRepository{}

		h := NewEmailChanged(m, idGen, vRepo)

		err := h.Handle(
			context.Background(),
			domainiam.UserVerifiedEvent{
				UserID:     newTestUserID(t, "user-1"),
				OccurredAt: time.Now(),
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if idGen.calls != 0 {
			t.Fatalf("expected no ID generation, got %d", idGen.calls)
		}

		if vRepo.addCalls != 0 {
			t.Fatalf("expected no verification add, got %d", vRepo.addCalls)
		}

		if m.sendVerificationEmailCalls != 0 {
			t.Fatalf(
				"expected no verification email, got %d",
				m.sendVerificationEmailCalls,
			)
		}
	})

	t.Run("propagates ID generation error", func(t *testing.T) {
		idGen := &fakeIDGenerator{
			err: errIDGeneration,
		}

		vRepo := &fakeVerificationRepository{}
		m := &fakeMailer{}

		h := NewEmailChanged(m, idGen, vRepo)

		err := h.Handle(
			context.Background(),
			domainiam.UserEmailChangedEvent{
				UserID:     newTestUserID(t, "user-1"),
				NewEmail:   newTestEmail(t, "new@example.com"),
				OccurredAt: time.Now(),
			},
		)

		if !errors.Is(err, errIDGeneration) {
			t.Fatalf("expected ID generation error, got %v", err)
		}

		if vRepo.addCalls != 0 {
			t.Fatalf("expected no verification add, got %d", vRepo.addCalls)
		}

		if m.sendVerificationEmailCalls != 0 {
			t.Fatalf(
				"expected no verification email, got %d",
				m.sendVerificationEmailCalls,
			)
		}
	})

	t.Run("propagates verification repository error", func(t *testing.T) {
		idGen := &fakeIDGenerator{
			id: "verification-1",
		}

		vRepo := &fakeVerificationRepository{
			addErr: errRepository,
		}

		m := &fakeMailer{}

		h := NewEmailChanged(m, idGen, vRepo)

		err := h.Handle(
			context.Background(),
			domainiam.UserEmailChangedEvent{
				UserID:     newTestUserID(t, "user-1"),
				NewEmail:   newTestEmail(t, "new@example.com"),
				OccurredAt: time.Now(),
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}

		if vRepo.addCalls != 1 {
			t.Fatalf("expected one verification add, got %d", vRepo.addCalls)
		}

		if m.sendVerificationEmailCalls != 0 {
			t.Fatalf(
				"expected no verification email, got %d",
				m.sendVerificationEmailCalls,
			)
		}
	})

	t.Run("propagates mailer error", func(t *testing.T) {
		idGen := &fakeIDGenerator{
			id: "verification-1",
		}

		vRepo := &fakeVerificationRepository{}

		m := &fakeMailer{
			sendVerificationEmailErr: errMailDelivery,
		}

		h := NewEmailChanged(m, idGen, vRepo)

		newEmail := newTestEmail(t, "new@example.com")

		err := h.Handle(
			context.Background(),
			domainiam.UserEmailChangedEvent{
				UserID:     newTestUserID(t, "user-1"),
				NewEmail:   newEmail,
				OccurredAt: time.Now(),
			},
		)

		if !errors.Is(err, errMailDelivery) {
			t.Fatalf("expected mailer error, got %v", err)
		}

		if vRepo.addCalls != 1 {
			t.Fatalf("expected one verification add, got %d", vRepo.addCalls)
		}

		if m.sendVerificationEmailCalls != 1 {
			t.Fatalf(
				"expected one verification email, got %d",
				m.sendVerificationEmailCalls,
			)
		}
	})

	t.Run("creates verification and sends email", func(t *testing.T) {
		idGen := &fakeIDGenerator{
			id: "verification-1",
		}

		vRepo := &fakeVerificationRepository{}
		m := &fakeMailer{}

		h := NewEmailChanged(m, idGen, vRepo)

		userID := newTestUserID(t, "user-1")
		newEmail := newTestEmail(t, "new@example.com")

		err := h.Handle(
			context.Background(),
			domainiam.UserEmailChangedEvent{
				UserID:     userID,
				NewEmail:   newEmail,
				OccurredAt: time.Now(),
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if idGen.calls != 1 {
			t.Fatalf("expected one ID generation, got %d", idGen.calls)
		}

		if vRepo.addCalls != 1 {
			t.Fatalf("expected one verification add, got %d", vRepo.addCalls)
		}

		if vRepo.verification == nil {
			t.Fatal("expected verification to be stored")
		}

		if vRepo.verification.UserID() != userID {
			t.Fatalf(
				"expected verification user ID %s, got %s",
				userID,
				vRepo.verification.UserID(),
			)
		}

		if vRepo.verification.ID().String() != "verification-1" {
			t.Fatalf(
				"expected verification ID verification-1, got %s",
				vRepo.verification.ID(),
			)
		}

		if m.sendVerificationEmailCalls != 1 {
			t.Fatalf(
				"expected one verification email, got %d",
				m.sendVerificationEmailCalls,
			)
		}

		if m.verificationEmail != newEmail {
			t.Fatalf(
				"expected email %s, got %s",
				newEmail,
				m.verificationEmail,
			)
		}

		if m.verificationID != vRepo.verification.ID() {
			t.Fatalf(
				"expected verification ID %s, got %s",
				vRepo.verification.ID(),
				m.verificationID,
			)
		}
	})
}
