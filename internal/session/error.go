package session

import "errors"

var (
	ErrNotFound    = errors.New("session: not found")
	ErrInvalidType = errors.New("session: invalid type")
)
