package eventhandler

import (
	"context"
	"errors"
	"testing"
	"time"

	domainiam "mizu/internal/domain/iam"
)

func TestSendVerifiedEmailHandle(t *testing.T) {
	t.Run("ignores unrelated event", func(t *testing.T) {
		userRepo := &fakeUserRepository{}
		m := &fakeMailer{}

		h := NewSendVerifiedEmail(m, userRepo)

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

		if m.sendVerifiedEmailCalls != 0 {
			t.Fatalf(
				"expected no verified email, got %d",
				m.sendVerifiedEmailCalls,
			)
		}
	})

	t.Run("propagates user repository error", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			getByIDErr: errRepository,
		}

		m := &fakeMailer{}

		h := NewSendVerifiedEmail(m, userRepo)

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

		if m.sendVerifiedEmailCalls != 0 {
			t.Fatalf(
				"expected no verified email, got %d",
				m.sendVerifiedEmailCalls,
			)
		}
	})

	t.Run("propagates mailer error", func(t *testing.T) {
		userID := newTestUserID(t, "user-1")

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
			sendVerifiedEmailErr: errMailDelivery,
		}

		h := NewSendVerifiedEmail(m, userRepo)

		err := h.Handle(
			context.Background(),
			domainiam.VerificationCreatedEvent{
				UserID:         userID,
				VerificationID: newTestVerificationID(t, "verification-1"),
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

		if m.sendVerifiedEmailCalls != 1 {
			t.Fatalf(
				"expected one verified email, got %d",
				m.sendVerifiedEmailCalls,
			)
		}

		if m.verifiedEmail != user.Email() {
			t.Fatalf(
				"expected verified email %s, got %s",
				user.Email(),
				m.verifiedEmail,
			)
		}
	})

	t.Run("sends verified email to user's email", func(t *testing.T) {
		userID := newTestUserID(t, "user-1")

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

		h := NewSendVerifiedEmail(m, userRepo)

		err := h.Handle(
			context.Background(),
			domainiam.VerificationCreatedEvent{
				UserID:         userID,
				VerificationID: newTestVerificationID(t, "verification-1"),
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

		if m.sendVerifiedEmailCalls != 1 {
			t.Fatalf(
				"expected one verified email, got %d",
				m.sendVerifiedEmailCalls,
			)
		}

		if m.verifiedEmail != user.Email() {
			t.Fatalf(
				"expected verified email %s, got %s",
				user.Email(),
				m.verifiedEmail,
			)
		}
	})
}
