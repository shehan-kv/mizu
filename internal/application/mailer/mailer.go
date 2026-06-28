package mailer

import (
	"context"
	"errors"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/verification"
)

var ErrMailDeliveryFailed = errors.New("failed to deliver email")

type Mailer interface {
	SendRecoveryEmail(ctx context.Context, email iam.Email, t iam.RecoveryToken) error
	SendVerificationEmail(ctx context.Context, email iam.Email, verificationID verification.VerificationID) error
	SendVerifiedEmail(ctx context.Context, email iam.Email) error
	SendContractEmail(ctx context.Context, email ContractEmail) error
	SendInvoiceEmail(ctx context.Context, email InvoiceEmail) error
}
