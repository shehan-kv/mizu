package eventhandler

import (
	"context"
	"errors"
	"testing"
	"time"

	domainiam "mizu/internal/domain/iam"
)

func TestSendVerificationEmailHandle(t *testing.T) {
	t.Run("ignores unrelated event", func(t *testing.T) {
		userRepo := &fakeUserRepository{}
		m := &fakeMailer{}

		h := NewSendVerificationEmail(m, userRepo)

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

		if userRepo.getByIDCalls != 0 {
			t.Fatalf(
				"expected no user lookup, got %d",
				userRepo.getByIDCalls,
			)
		}

		if m.sendVerificationEmailCalls != 0 {
			t.Fatalf(
				"expected no verification email, got %d",
				m.sendVerificationEmailCalls,
			)
		}
	})

	t.Run("propagates user repository error", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			getByIDErr: errRepository,
		}

		m := &fakeMailer{}

		h := NewSendVerificationEmail(m, userRepo)

		err := h.Handle(
			context.Background(),
			domainiam.VerificationCreatedEvent{
				UserID:         newTestUserID(t, "user-1"),
				VerificationID: newTestVerificationID(t, "verification-1"),
				OccurredAt:     time.Now(),
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf(
				"expected repository error, got %v",
				err,
			)
		}

		if userRepo.getByIDCalls != 1 {
			t.Fatalf(
				"expected one user lookup, got %d",
				userRepo.getByIDCalls,
			)
		}

		if m.sendVerificationEmailCalls != 0 {
			t.Fatalf(
				"expected no verification email, got %d",
				m.sendVerificationEmailCalls,
			)
		}
	})

	t.Run("propagates mailer error", func(t *testing.T) {
		userID := newTestUserID(t, "user-1")
		verificationID := newTestVerificationID(
			t,
			"verification-1",
		)

		user := newTestUser(
			t,
			"user-1",
			domainiam.RoleClient,
			true,
		)

		userRepo := &fakeUserRepository{
			user: user,
		}

		m := &fakeMailer{
			sendVerificationEmailErr: errMailDelivery,
		}

		h := NewSendVerificationEmail(m, userRepo)

		err := h.Handle(
			context.Background(),
			domainiam.VerificationCreatedEvent{
				UserID:         userID,
				VerificationID: verificationID,
				OccurredAt:     time.Now(),
			},
		)

		if !errors.Is(err, errMailDelivery) {
			t.Fatalf(
				"expected mailer error, got %v",
				err,
			)
		}

		if userRepo.getByIDCalls != 1 {
			t.Fatalf(
				"expected one user lookup, got %d",
				userRepo.getByIDCalls,
			)
		}

		if m.sendVerificationEmailCalls != 1 {
			t.Fatalf(
				"expected one verification email, got %d",
				m.sendVerificationEmailCalls,
			)
		}

		if m.verificationEmail != user.Email() {
			t.Fatalf(
				"expected verification email %s, got %s",
				user.Email(),
				m.verificationEmail,
			)
		}

		if m.verificationID != verificationID {
			t.Fatalf(
				"expected verification ID %s, got %s",
				verificationID,
				m.verificationID,
			)
		}
	})

	t.Run("sends verification email to user's email", func(t *testing.T) {
		userID := newTestUserID(t, "user-1")
		verificationID := newTestVerificationID(
			t,
			"verification-1",
		)

		user := newTestUser(
			t,
			"user-1",
			domainiam.RoleClient,
			true,
		)

		userRepo := &fakeUserRepository{
			user: user,
		}

		m := &fakeMailer{}

		h := NewSendVerificationEmail(m, userRepo)

		err := h.Handle(
			context.Background(),
			domainiam.VerificationCreatedEvent{
				UserID:         userID,
				VerificationID: verificationID,
				OccurredAt:     time.Now(),
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if userRepo.getByIDCalls != 1 {
			t.Fatalf(
				"expected one user lookup, got %d",
				userRepo.getByIDCalls,
			)
		}

		if m.sendVerificationEmailCalls != 1 {
			t.Fatalf(
				"expected one verification email, got %d",
				m.sendVerificationEmailCalls,
			)
		}

		if m.verificationEmail != user.Email() {
			t.Fatalf(
				"expected verification email %s, got %s",
				user.Email(),
				m.verificationEmail,
			)
		}

		if m.verificationID != verificationID {
			t.Fatalf(
				"expected verification ID %s, got %s",
				verificationID,
				m.verificationID,
			)
		}
	})
}
