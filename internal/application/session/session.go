package session

import (
	"mizu/internal/domain/iam"
	"time"
)

type Session struct {
	id        SessionID
	userID    iam.UserID
	issuedAt  time.Time
	expiresAt time.Time
}

func NewSession(id SessionID, userID iam.UserID) *Session {
	return &Session{
		id:        id,
		userID:    userID,
		issuedAt:  time.Now(),
		expiresAt: time.Now().Add(72 * time.Hour),
	}
}

func (s *Session) ID() SessionID {
	return s.id
}

func (s *Session) UserID() iam.UserID {
	return s.userID
}

func (s *Session) IssuedAt() time.Time {
	return s.issuedAt
}

func (s *Session) ExpiresAt() time.Time {
	return s.expiresAt
}

func (s *Session) Equals(session Session) bool {
	return s.id == session.ID()
}
