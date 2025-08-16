package sse

// Represents a client in SSE.
//
// Id is the internal user-Id used by the system
// to identify a user ( eg: database assigned user Id ).
//
// Any event for this user should be added to the SendQueue
// channel
type Client struct {
	Id        int64
	SendQueue chan []byte
}
