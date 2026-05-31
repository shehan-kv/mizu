package sse

// Represents a client in SSE.
//
// Id is the internal user-Id used by the system
// to identify a user.
//
// Any event for this user should be added to the SendQueue
// channel
type Client struct {
	Id        string
	SendQueue chan []byte
}
