package session

import (
	"sync"
	"time"
)

// InMemorySession is an implementation of the
// [/internal/session/SessionStore] interface
// Uses a sync.Map (thread-safe) to store sessions in-memory.
type InMemorySession struct {
	sessions sync.Map
}

// NewInMemeorySession creates a new instance of InMemorySession.
// Returns:
//   - a pointer to a [/internal/session/InMemorySession] struct
func NewInMemorySession() *InMemorySession {
	return &InMemorySession{}
}

// Not-implemented.
// Not required for the in-memory session management
// Added to comply with the interface
func (ims *InMemorySession) Init() {
}

// Adds a session to the in-memory store
//
// Parameters:
//   - key: a string that uniquely identifies a session (ex: a UUID)
//   - userId: the id of the user to store in the session
//
// Returns:
//   - nil as error for this implementation, for compliance with the interface
func (ims *InMemorySession) SetSession(key string, userId int64) error {

	ims.sessions.Store(key, Session{UserId: userId, IssuedAt: time.Now()})
	return nil
}

// Gets a session from the in-memory store
//
// Parameters:
//   - key: a string that uniquely identifies a session (ex: a UUID)
//
// Returns:
//   - *Session: if session is found
//   - ErrNotFound: if session not found
//   - ErrInvalidType: if type is invalid
func (ims *InMemorySession) GetSession(key string) (*Session, error) {

	value, ok := ims.sessions.Load(key)
	if !ok {
		return &Session{}, ErrNotFound
	}

	session, ok := value.(Session)
	if !ok {
		return &Session{}, ErrInvalidType
	}

	return &session, nil
}

// Revokes a session using a session key.
// This function is idempotent.
//
// Parameters:
//   - key: a string that uniquely identifies a session (ex: a UUID)
//
// Returns:
//   - nil, this implementation doesn't cause errors
func (ims *InMemorySession) RevokeSession(key string) error {

	ims.sessions.Delete(key)
	return nil
}

// Not-implemented.
// Not required for the in-memory session management
// Added to comply with the interface
func (ims *InMemorySession) Close() {

}
