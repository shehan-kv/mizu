package email

import "mizu/internal/logger"

// GetEmailSender chooses an email sender EMAIL environment variable
// Parameters:
//   - lg: a logger that implements the logger.Logger interface
//
// Returns:
//   - an implementation of the EmailSender
func GetEmailSender(lg logger.Logger) EmailSender {

	// Dynamically choose the email sender
	// NoOpEmailSender is always used until the other
	// senders are implemented
	return NewNoOpEmailSender(lg)
}
