package store

import "errors"

var (
	ErrInsertFailed        = errors.New("database: insert failed")
	ErrUpdateFailed        = errors.New("database: update failed")
	ErrDeleteFailed        = errors.New("database: delete failed")
	ErrUniqueViolation     = errors.New("database: unique field violation")
	ErrForeignKeyViolation = errors.New("database: foreign key violation")
	ErrNotNullViolation    = errors.New("database: not null violation")
	ErrCheckViolation      = errors.New("database: not null violation")
	ErrRecordNotFound      = errors.New("database: record not found")
	ErrQueryFailed         = errors.New("database: query failed")
	ErrUnexpectedType      = errors.New("database: unexpected type")
)
