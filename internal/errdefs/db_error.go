package errdefs

import "errors"

// Database errors
var (
	ErrDbInsertFailed    = errors.New("database: insert failed")
	ErrDbUpdateFailed    = errors.New("database: update failed")
	ErrDbDeleteFailed    = errors.New("database: delete failed")
	ErrDbRecordNotFound  = errors.New("database: record not found")
	ErrDbUniqueViolation = errors.New("database: unique field violation")
	ErrDbQueryFailed     = errors.New("database: query failed")
)
