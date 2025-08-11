package email

// GetEmailSender chooses an email sender EMAIL environment variable
//
// Returns:
//   - an implementation of the EmailSender
func GetEmailSender() EmailSender {

	// Dynamically choose the email sender
	// NoOpEmailSender is always used until the other
	// senders are implemented
	return NewNoOpEmailSender()
}
