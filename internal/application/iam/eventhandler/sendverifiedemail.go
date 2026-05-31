package eventhandler

import (
	"context"
	"mizu/internal/application/mailer"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/verification"
)

type SendVerifiedEmail struct {
	mailer  mailer.Mailer
	iamRepo iam.Repository
}

func NewSendVerifiedEmail(mailer mailer.Mailer, iamRepo iam.Repository) *SendVerifiedEmail {
	return &SendVerifiedEmail{mailer: mailer, iamRepo: iamRepo}
}

func (h *SendVerifiedEmail) Handle(ctx context.Context, event common.Event) error {

	e, ok := event.(verification.VerificationCreatedEvent)
	if !ok {
		return nil
	}

	u, err := h.iamRepo.GetByID(ctx, e.UserID)
	if err != nil {
		return err
	}

	return h.mailer.SendVerifiedEmail(ctx, u.Email())
}
