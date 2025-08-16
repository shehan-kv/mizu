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

// Adds a new client.
// Replaces the client if already exists.
// This method is thread-safe.
//
// Parameters:
//   - id: a unique identifier per connected client
//   - c: a pointer to a *Client
func (s *SseSender) AddClient(connId string, userId int64) <-chan []byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	client := &Client{
		Id:        userId,
		SendQueue: make(chan []byte, 100),
	}

	s.clients[connId] = client

	if s.byClient[userId] == nil {
		s.byClient[userId] = make(map[string]*Client)
	}

	s.byClient[userId][connId] = client

	return client.SendQueue
}
