package sse

import (
	"sync"
)

// Keeps track of connected clients.
// Handles adding, removing and sending messages to
// clients in a thread-safe way.
type SseSender struct {

	// a map of unique connection Id to *Client
	// for efficient broadcasting purposes
	clients map[string]*Client

	// map of user-Id to unique connection Ids to *Client
	// for efficiently sending a message to a few selected
	// clients
	byClient map[int64]map[string]*Client
	mu       sync.RWMutex
}

// Creates a new instance of SseSender.
//
// Returns:
//   - a pointer to new SseSender
func NewSseSender() *SseSender {
	return &SseSender{
		clients:  make(map[string]*Client),
		byClient: make(map[int64]map[string]*Client),
	}
}
