package file

import "errors"

var (
	ErrDirCreateFailed  = errors.New("file: failed to create directory")
	ErrFileCreateFailed = errors.New("file: failed to create file")
	ErrCopyFailed       = errors.New("file: failed to copy data")
	ErrFileCloseFailed  = errors.New("file: failed to close file")
	ErrRenameFailed     = errors.New("file: failed to rename")
)
