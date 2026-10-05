package common

import "errors"

var (
	ErrFailedToGenerateID = errors.New("failed to generate id")
	ErrInvalidPageLimit   = errors.New("invalid page limit")
	ErrPageLimitExceeded  = errors.New("page limit exceeded")
	ErrInvalidPageOffset  = errors.New("invalid page offset")
)
