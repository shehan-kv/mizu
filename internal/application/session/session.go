package session

import (
	"encoding/json"
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

func (s Session) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID        SessionID  `json:"id"`
		UserID    iam.UserID `json:"user_id"`
		IssuedAt  time.Time  `json:"issued_at"`
		ExpiresAt time.Time  `json:"expires_at"`
	}{
		ID:        s.id,
		UserID:    s.userID,
		IssuedAt:  s.issuedAt,
		ExpiresAt: s.expiresAt,
	})
}

func (s *Session) UnmarshalJSON(data []byte) error {
	var v struct {
		ID        SessionID  `json:"id"`
		UserID    iam.UserID `json:"user_id"`
		IssuedAt  time.Time  `json:"issued_at"`
		ExpiresAt time.Time  `json:"expires_at"`
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	s.id = v.ID
	s.userID = v.UserID
	s.issuedAt = v.IssuedAt
	s.expiresAt = v.ExpiresAt

	return nil
}
