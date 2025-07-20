package errdefs

import "errors"

// Session errors
var (
	ErrSessionNotFound     = errors.New("session: not found")
	ErrSessionRevokeFailed = errors.New("session: revoke failed")
	ErrSessionSetFailed    = errors.New("session: set failed")
	ErrSessionInvalidType  = errors.New("session: invalid type")
)
