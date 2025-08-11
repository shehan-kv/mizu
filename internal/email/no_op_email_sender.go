package email

import (
	"context"
	"mizu/internal/email/params"
	"mizu/internal/event"
	"mizu/internal/logger"
)

type NoOpEmailSender struct {
	lg logger.Logger
}

// Creates a new instance of NoOpEmailSender.
//
// Parameters:
//   - lg: an implementation of logger.Logger interface
//
// Returns:
//   - a pointer to a NoOpEmailSender struct
func NewNoOpEmailSender(lg logger.Logger) *NoOpEmailSender {
	return &NoOpEmailSender{lg: lg}
}

// Not-implemented.
// Not required for the no-op email sender
// Added to comply with the interface
func (s *NoOpEmailSender) Init() {
}

// Impementation of SendVerifyRequest of EmailSender interface
// Logs a WARN message with correlation ID
func (s *NoOpEmailSender) SendVerifyRequest(ctx context.Context, arg *params.VerifyRequest) error {

	s.lg.Warn("no-op email sender configured, onboarding email not sent",
		"event", event.EventEmailSendFailed,
		"scope", "no_op_email_sender",
		"correlation_id", arg.CorrelationId)
	return nil
}

// Not-implemented.
// Not required for the no-op email sender
// Added to comply with the interface
func (s *NoOpEmailSender) Close() {
}
