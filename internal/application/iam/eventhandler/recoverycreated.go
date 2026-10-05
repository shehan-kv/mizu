package eventhandler

import (
	"context"
	"mizu/internal/application/mailer"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
)

type RecoveryCreated struct {
	mailer   mailer.Mailer
	userRepo iam.UserRepository
}

func NewRecoveryCreated(mailer mailer.Mailer, userRepo iam.UserRepository) *RecoveryCreated {
	return &RecoveryCreated{
		mailer:   mailer,
		userRepo: userRepo,
	}
}

func (h *RecoveryCreated) Handle(ctx context.Context, event common.Event) error {

	e, ok := event.(iam.RecoveryCreatedEvent)
	if !ok {
		return nil
	}

	u, err := h.userRepo.GetByID(ctx, e.UserID)
	if err != nil {
		return err
	}

	return h.mailer.SendRecoveryEmail(ctx, u.Email(), e.Token)
}
