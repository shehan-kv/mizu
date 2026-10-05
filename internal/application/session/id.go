package session

type SessionID string

func NewSessionID(id string) (SessionID, error) {
	if id == "" {
		return "", ErrSessionIDCannotBeEmpty
	}

	return SessionID(id), nil
}

func (s SessionID) String() string {
	return string(s)
}
