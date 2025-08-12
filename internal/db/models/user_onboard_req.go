package models

import "time"

type UserOnboardReq struct {
	Id       int64
	UserId   int64
	Token    string
	IssuedAt time.Time
	IsValid  bool
}
