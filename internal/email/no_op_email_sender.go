package email

import (
	"context"
	"mizu/internal/email/params"
)

type NoOpEmailSender struct {
}

// Creates a new instance of NoOpEmailSender.
//
// Returns:
//   - a pointer to a NoOpEmailSender struct
func NewNoOpEmailSender() *NoOpEmailSender {
	return &NoOpEmailSender{}
}

// Not-implemented.
// Not required for the no-op email sender
// Added to comply with the interface
func (s *NoOpEmailSender) Init() {
}

// Impementation of SendVerifyRequest of EmailSender interface
// Logs a WARN message with correlation ID
func (s *NoOpEmailSender) SendVerifyRequest(ctx context.Context, arg *params.VerifyRequest) error {

	return ErrEmailSendFailed
}

func (s *NoOpEmailSender) SendContractSigned(ctx context.Context, arg *params.ContractSignedRequest) error {

	return ErrEmailSendFailed
}

// Not-implemented.
// Not required for the no-op email sender
// Added to comply with the interface
func (s *NoOpEmailSender) Close() {
}
