package session

import (
	"context"
	"mizu/internal/application/session"
	"mizu/internal/domain/iam"
	"sync"
)

// InMemoryStore is a thread-safe in-memory implementation of Store.
type InMemoryStore struct {
	mu             sync.RWMutex
	sessions       map[session.SessionID]*session.Session
	userSessionIDs map[iam.UserID][]session.SessionID
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		sessions:       make(map[session.SessionID]*session.Session),
		userSessionIDs: make(map[iam.UserID][]session.SessionID),
	}
}

// Add stores the session in the store.
func (s *InMemoryStore) Add(ctx context.Context, sess *session.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sess.ID()] = sess
	s.userSessionIDs[sess.UserID()] = append(s.userSessionIDs[sess.UserID()], sess.ID())
	return nil
}

// Get retrieves a session by its ID.
// Returns ErrSessionNotFound if the session does not exist.
func (s *InMemoryStore) Get(ctx context.Context, id session.SessionID) (*session.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, session.ErrSessionNotFound
	}
	return sess, nil
}

// ListByUser returns all active sessions for the given user.
// Returns nil if the user has no active sessions.
func (s *InMemoryStore) ListByUser(ctx context.Context, userID iam.UserID) ([]*session.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids, ok := s.userSessionIDs[userID]
	if !ok {
		return nil, nil
	}
	sessions := make([]*session.Session, 0, len(ids))
	for _, id := range ids {
		if sess, ok := s.sessions[id]; ok {
			sessions = append(sessions, sess)
		}
	}
	return sessions, nil
}

// Delete removes a session by its ID.
// Delete is idempotent — deleting a non-existent session is not an error.
func (s *InMemoryStore) Delete(ctx context.Context, id session.SessionID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[id]
	if !ok {
		return nil
	}

	delete(s.sessions, id)

	userID := sess.UserID()
	ids := s.userSessionIDs[userID]
	for i, sID := range ids {
		if sID == id {
			s.userSessionIDs[userID] = append(ids[:i], ids[i+1:]...)
			break
		}
	}

	if len(s.userSessionIDs[userID]) == 0 {
		delete(s.userSessionIDs, userID)
	}

	return nil
}
