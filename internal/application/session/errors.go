package session

import "errors"

var (
	ErrSessionIDCannotBeEmpty = errors.New("session id cannot be empty")
	ErrSessionNotFound        = errors.New("session not found")
)
