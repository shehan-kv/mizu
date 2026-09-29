package email

import (
	"context"
	"testing"

	"mizu/internal/application/logger"
	"mizu/internal/application/mailer"
	"mizu/internal/domain/iam"
)

type fakeLogger struct{}

func (fakeLogger) Info(string, ...any)  {}
func (fakeLogger) Warn(string, ...any)  {}
func (fakeLogger) Error(string, ...any) {}
func (fakeLogger) Fatal(string, ...any) {}

var _ logger.Logger = (*fakeLogger)(nil)

func TestNoOpMailer(t *testing.T) {
	m := NewNoOpMailer(fakeLogger{})
	ctx := context.Background()

	email, err := iam.NewEmail("test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	recoveryToken, err := iam.NewRecoveryToken("recovery-token")
	if err != nil {
		t.Fatal(err)
	}

	verificationID, err := iam.NewVerificationID("verification-id")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		send func() error
	}{
		{
			name: "recovery email",
			send: func() error {
				return m.SendRecoveryEmail(
					ctx,
					email,
					recoveryToken,
				)
			},
		},
		{
			name: "verification email",
			send: func() error {
				return m.SendVerificationEmail(
					ctx,
					email,
					verificationID,
				)
			},
		},
		{
			name: "verified email",
			send: func() error {
				return m.SendVerifiedEmail(
					ctx,
					email,
				)
			},
		},
		{
			name: "contract email",
			send: func() error {
				return m.SendContractEmail(
					ctx,
					mailer.ContractEmail{},
				)
			},
		},
		{
			name: "invoice email",
			send: func() error {
				return m.SendInvoiceEmail(
					ctx,
					mailer.InvoiceEmail{},
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.send(); err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}
