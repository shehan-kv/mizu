package email

import (
	"context"
	"mizu/internal/email/params"
)

type EmailSender interface {
	Init()
	SendVerifyRequest(ctx context.Context, arg *params.VerifyRequest) error
	Close()
}
