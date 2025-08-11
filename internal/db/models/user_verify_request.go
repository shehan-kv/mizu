package models

import "time"

type UserVerifyRequest struct {
	Id       int64
	UserId   int64
	Token    string
	IssuedAt time.Time
	IsValid  bool
}
