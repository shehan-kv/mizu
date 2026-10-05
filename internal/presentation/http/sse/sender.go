package sse

import (
	"sync"
)

// Sender keeps track of connected clients.
// Handles adding, removing and sending messages to
// clients in a thread-safe way.
type Sender struct {

	// a map of unique connection Id to *Client
	// for efficient broadcasting purposes
	clients map[string]*Client

	// map of user-Id to unique connection Ids to *Client
	// for efficiently sending a message to a few selected
	// clients
	byClient map[string]map[string]*Client
	mu       sync.RWMutex
}

// Creates a new instance of SseSender.
//
// Returns:
//   - a pointer to new SseSender
func NewSender() *Sender {
	return &Sender{
		clients:  make(map[string]*Client),
		byClient: make(map[string]map[string]*Client),
	}
}

// Adds a new client.
// Replaces the client if already exists.
// This method is thread-safe.
func (s *Sender) AddClient(connId string, userId string) <-chan []byte {
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

// RemoveClient removes a connected client.
// This method if thread-safe.
//
// id is the connected client's unique id and
// userId is the id of the user
func (s *Sender) RemoveClient(id string, userId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, ok := s.clients[id]
	if ok {
		close(client.SendQueue)
	}

	delete(s.clients, id)
	delete(s.byClient[userId], id)
}

// SendTo sends a message to a list of specific clients
// identified by their user-Ids.
// This method if thread-safe.
func (s *Sender) SendTo(event string, msg []byte, to []string) {

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

// SendAll broadcasts a message to all connected users.
// This method if thread-safe.
func (s *Sender) SendAll(event string, msg []byte) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		client.SendQueue <- buildMessage(event, msg)
	}

}

func (s *Sender) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, c := range s.clients {
		close(c.SendQueue)
	}

	s.clients = map[string]*Client{}
	s.byClient = map[string]map[string]*Client{}
}

// buildMessage builds the server sent event message
func buildMessage(event string, msg []byte) []byte {

	bytesToAllocate := len("event: \n") + len(event) + len("data: \n\n") + len(msg)

	buf := make([]byte, 0, bytesToAllocate)

	buf = append(buf, "event: "...)
	buf = append(buf, event...)
	buf = append(buf, '\n')
	buf = append(buf, "data: "...)
	buf = append(buf, msg...)
	buf = append(buf, '\n', '\n')

	return buf
}
