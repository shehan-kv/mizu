package iam

import "time"

type SessionDTO struct {
	ID        string
	UserID    string
	Role      string
	IssuedAt  time.Time
	ExpiresAt time.Time
}
