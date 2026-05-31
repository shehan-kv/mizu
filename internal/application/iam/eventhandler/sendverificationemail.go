package eventhandler

import (
	"context"
	"mizu/internal/application/mailer"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/verification"
)

type SendVerificationEmail struct {
	mailer  mailer.Mailer
	iamRepo iam.Repository
}

func NewSendVerificationEmail(mailer mailer.Mailer, iamRepo iam.Repository) *SendVerificationEmail {
	return &SendVerificationEmail{mailer: mailer, iamRepo: iamRepo}
}

func (h *SendVerificationEmail) Handle(ctx context.Context, event common.Event) error {

	e, ok := event.(verification.VerificationCreatedEvent)
	if !ok {
		return nil
	}

	u, err := h.iamRepo.GetByID(ctx, e.UserID)
	if err != nil {
		return err
	}

	return h.mailer.SendVerificationEmail(ctx, u.Email(), e.VerificationID)
}
