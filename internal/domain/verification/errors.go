package verification

import "errors"

var (
	ErrVerificationIDCannotBeEmpty        = errors.New("verification id cannot be empty")
	ErrVerificationUserIDCannotBeEmpty    = errors.New("verification user id cannot be empty")
	ErrVerificationNotFound               = errors.New("verification not found")
	ErrVerificationConcurrentModification = errors.New("verification concurrent modification")
)
