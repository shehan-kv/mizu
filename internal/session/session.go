package session

import "time"

// Session represents a user session
type Session struct {
	UserId   int64
	IssuedAt time.Time
}
