package service

import "errors"

var (
	ErrUnauthorized  = errors.New("service: unauthorized")
	ErrInternalError = errors.New("service: internal error")
	ErrAlreadyExists = errors.New("service: already exists")
	ErrBadRequest    = errors.New("service: bad request")
	ErrNotFound      = errors.New("service: not found")
)
