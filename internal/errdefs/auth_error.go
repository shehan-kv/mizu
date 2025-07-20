package errdefs

import "errors"

// Authentication errors
var (
	ErrAuthUnauthorized  = errors.New("auth: unauthorized")
	ErrAuthInternalError = errors.New("auth: internal error")
)
