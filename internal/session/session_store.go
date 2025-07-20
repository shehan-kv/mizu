package session

// Defines the behavior required for
// managing user sessions.
type SessionStore interface {
	Init()
	SetSession(key string, userId int64) error
	GetSession(key string) (*Session, error)
	RevokeSession(key string) error
	Close()
}
