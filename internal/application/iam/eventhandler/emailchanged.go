package eventhandler

import (
	"context"
	"mizu/internal/application/mailer"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/verification"
	"time"
)

type EmailChanged struct {
	mailer mailer.Mailer
	idGen  common.IDGenerator
	vRepo  verification.Repository
}

func NewEmailChanged(
	mailer mailer.Mailer,
	idGen common.IDGenerator,
	vRepo verification.Repository,
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

	vID, err := verification.NewVerificationID(id)
	if err != nil {
		return err
	}

	v := verification.NewVerification(vID, e.UserID, time.Now())

	err = h.vRepo.Add(ctx, v)
	if err != nil {
		return err
	}

	return h.mailer.SendVerificationEmail(ctx, e.NewEmail, vID)
}
