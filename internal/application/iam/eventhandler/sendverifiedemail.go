package eventhandler

import (
	"context"
	"mizu/internal/application/mailer"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
)

type SendVerifiedEmail struct {
	mailer   mailer.Mailer
	userRepo iam.UserRepository
}

func NewSendVerifiedEmail(mailer mailer.Mailer, userRepo iam.UserRepository) *SendVerifiedEmail {
	return &SendVerifiedEmail{mailer: mailer, userRepo: userRepo}
}

func (h *SendVerifiedEmail) Handle(ctx context.Context, event common.Event) error {

	e, ok := event.(iam.VerificationCreatedEvent)
	if !ok {
		return nil
	}

	u, err := h.userRepo.GetByID(ctx, e.UserID)
	if err != nil {
		return err
	}

	return h.mailer.SendVerifiedEmail(ctx, u.Email())
}
