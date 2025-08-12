package models

import "time"

type UserOnboardRequest struct {
	Id       int64
	UserId   int64
	Token    string
	IssuedAt time.Time
	IsValid  bool
}
