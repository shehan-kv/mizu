package email

import (
	"context"
	"mizu/internal/application/logger"
	"mizu/internal/application/mailer"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/verification"
)

type NoOpMailer struct {
	logger logger.Logger
}

func NewNoOpMailer(logger logger.Logger) *NoOpMailer {
	return &NoOpMailer{logger: logger}
}

func (m *NoOpMailer) SendVerificationEmail(_ context.Context, _ iam.Email, _ verification.VerificationID) error {

	m.logger.Info("verification email skipped: using no-op mailer")

	return nil
}

func (m *NoOpMailer) SendVerifiedEmail(_ context.Context, _ iam.Email) error {
	m.logger.Info("verified email skipped: using no-op mailer")

	return nil
}

func (m *NoOpMailer) SendContractEmail(ctx context.Context, email mailer.ContractEmail) error {
	m.logger.Info("contract email skipped: using no-op mailer")

	return nil
}
