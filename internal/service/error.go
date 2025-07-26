package service

import "errors"

var (
	ErrUnauthorized  = errors.New("service: unauthorized")
	ErrInternalError = errors.New("service: internal error")
	ErrAlreadyExists = errors.New("service: already exists")
)
