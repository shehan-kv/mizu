package eventhandler

import (
	"context"
	"errors"
	"testing"
	"time"

	domainiam "mizu/internal/domain/iam"
)

func TestRecoveryCreatedHandle(t *testing.T) {
	t.Run("ignores unrelated event", func(t *testing.T) {
		userRepo := &fakeUserRepository{}
		m := &fakeMailer{}

		h := NewRecoveryCreated(m, userRepo)

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

		if m.sendRecoveryEmailCalls != 0 {
			t.Fatalf(
				"expected no recovery email, got %d",
				m.sendRecoveryEmailCalls,
			)
		}
	})

	t.Run("propagates user repository error", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			getByIDErr: errRepository,
		}

		m := &fakeMailer{}

		h := NewRecoveryCreated(m, userRepo)

		userID := newTestUserID(t, "user-1")
		token := newTestRecoveryToken(t, "recovery-token")

		err := h.Handle(
			context.Background(),
			domainiam.RecoveryCreatedEvent{
				UserID:     userID,
				Token:      token,
				OccurredAt: time.Now(),
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

		if m.sendRecoveryEmailCalls != 0 {
			t.Fatalf(
				"expected no recovery email, got %d",
				m.sendRecoveryEmailCalls,
			)
		}
	})

	t.Run("propagates mailer error", func(t *testing.T) {
		userID := newTestUserID(t, "user-1")
		email := newTestEmail(t, "user@example.com")
		token := newTestRecoveryToken(t, "recovery-token")

		user := newTestUser(
			t,
			userID.String(),
			domainiam.RoleClient,
			true,
		)

		userRepo := &fakeUserRepository{
			user: user,
		}

		m := &fakeMailer{
			sendRecoveryEmailErr: errMailDelivery,
		}

		h := NewRecoveryCreated(m, userRepo)

		err := h.Handle(
			context.Background(),
			domainiam.RecoveryCreatedEvent{
				UserID:     userID,
				Token:      token,
				OccurredAt: time.Now(),
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

		if m.sendRecoveryEmailCalls != 1 {
			t.Fatalf(
				"expected one recovery email, got %d",
				m.sendRecoveryEmailCalls,
			)
		}

		if m.recoveryEmail != email {
			t.Fatalf(
				"expected recovery email %s, got %s",
				email,
				m.recoveryEmail,
			)
		}

		if m.recoveryToken != token {
			t.Fatalf(
				"expected recovery token %s, got %s",
				token,
				m.recoveryToken,
			)
		}
	})

	t.Run("sends recovery email to user's email", func(t *testing.T) {
		userID := newTestUserID(t, "user-1")
		token := newTestRecoveryToken(t, "recovery-token")

		user := newTestUser(
			t,
			userID.String(),
			domainiam.RoleClient,
			true,
		)

		userRepo := &fakeUserRepository{
			user: user,
		}

		m := &fakeMailer{}

		h := NewRecoveryCreated(m, userRepo)

		err := h.Handle(
			context.Background(),
			domainiam.RecoveryCreatedEvent{
				UserID:     userID,
				Token:      token,
				OccurredAt: time.Now(),
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

		if m.sendRecoveryEmailCalls != 1 {
			t.Fatalf(
				"expected one recovery email, got %d",
				m.sendRecoveryEmailCalls,
			)
		}

		if m.recoveryEmail != user.Email() {
			t.Fatalf(
				"expected recovery email %s, got %s",
				user.Email(),
				m.recoveryEmail,
			)
		}

		if m.recoveryToken != token {
			t.Fatalf(
				"expected recovery token %s, got %s",
				token,
				m.recoveryToken,
			)
		}
	})
}
