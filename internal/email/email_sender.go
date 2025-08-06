package email

import (
	"context"
	"mizu/internal/email/params"
)

type EmailSender interface {
	Init()
	SendOnboardingRequest(ctx context.Context, arg *params.OnboardingRequest) error
	Close()
}
