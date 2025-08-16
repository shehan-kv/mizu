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

// Removes client.
// This method if thread-safe.
//
// Parameters:
//   - id: connected client's unique id
//   - userId: id of the user
func (s *SseSender) RemoveClient(id string, userId int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, id)
	delete(s.byClient[userId], id)
}

// Send a message to a list of specific clients
// identified by their user-Ids.
// This method if thread-safe.
//
// Parameters:
//   - event: event name
//   - msg: message to send
//   - to: a list of user-Ids to send the message to
func (s *SseSender) SendTo(event string, msg []byte, to []int64) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, id := range to {

		connections, ok := s.byClient[id]
		if !ok {
			continue
		}

		for _, client := range connections {
			client.SendQueue <- buildMessage(event, msg)
		}

	}

}

// Builds the server sent event message
//
// Parameters:
//   - event: event name
//   - msg: message to send
//
// Returns:
//   - []byte: server sent event to send
func buildMessage(event string, msg []byte) []byte {

	bytesToAllocate := len("event: \n") + len(event) + len("data: \n\n") + len(msg)

	buf := make([]byte, bytesToAllocate)

	buf = append(buf, "event: "...)
	buf = append(buf, event...)
	buf = append(buf, '\n')
	buf = append(buf, "data: "...)
	buf = append(buf, msg...)
	buf = append(buf, '\n', '\n')

	return buf
}
