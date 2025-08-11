package email

import "errors"

var (
	ErrEmailSendFailed error = errors.New("email: email send failed")
)
