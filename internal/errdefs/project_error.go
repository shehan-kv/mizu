package errdefs

import "errors"

// Project errors
var (
	ErrAlreadyExists        = errors.New("project: already exists")
	ErrProjectInternalError = errors.New("project: internal error")
	ErrContextDataNotFound  = errors.New("context: context data not found")
)
