package eventhandler

import (
	"context"
	"mizu/internal/application/mailer"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"time"
)

type EmailChanged struct {
	mailer mailer.Mailer
	idGen  common.IDGenerator
	vRepo  iam.VerificationRepository
}

func NewEmailChanged(
	mailer mailer.Mailer,
	idGen common.IDGenerator,
	vRepo iam.VerificationRepository,
) *EmailChanged {
	return &EmailChanged{
		mailer: mailer,
		idGen:  idGen,
		vRepo:  vRepo,
	}
}

func (h *EmailChanged) Handle(ctx context.Context, event common.Event) error {

	e, ok := event.(iam.UserEmailChangedEvent)
	if !ok {
		return nil
	}

	id, err := h.idGen.Generate()
	if err != nil {
		return err
	}

	vID, err := iam.NewVerificationID(id)
	if err != nil {
		return err
	}

	v := iam.NewVerification(vID, e.UserID, time.Now())

	err = h.vRepo.Add(ctx, v)
	if err != nil {
		return err
	}

	return h.mailer.SendVerificationEmail(ctx, e.NewEmail, vID)
}
