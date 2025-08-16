package event

import (
	"mizu/internal/sse"
)

// Sends events to clients.
//
// If a message queue is configured, posts the message
// to the message queue.
// Otherwise uses SSE.
type EventSender struct {
	sseSndr *sse.SseSender
}

// Creates a new instance of EventSender.
//
// Parameters:
//   - sseSndr: a pointer to a sse.SseSender
//
// Returns:
//   - a pointer to new EventSender
func NewEventSender(sseSndr *sse.SseSender) *EventSender {
	return &EventSender{
		sseSndr: sseSndr,
	}
}

// Send a message to a list of specific clients
// identified by their user-Ids.
//
// Parameters:
//   - event: event name
//   - msg: message to send
//   - to: a list of user-Ids to send the message to
func (n *EventSender) SendTo(event string, msg []byte, to []int64) {

	n.sseSndr.SendTo(event, msg, to)
}

// Broadcasts a message to all connected users.
// This method if thread-safe.
//
// Parameters:
//   - event: event name
//   - msg: message to send
func (n *EventSender) SendToAll(event string, msg []byte) {

	n.sseSndr.SendAll(event, msg)
}
