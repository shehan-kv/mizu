package eventbus

import (
	"context"
)

// EventType identifies the kind of event being published.
type EventType string

// Event is the contract for all events published through the bus.
type Event interface {
	EventType() EventType
}

type ExtEventHandler func(ctx context.Context, event Event) error

// ExternalBus delivers events across all running instances.
// Used for real-time notifications, SSE push, and cross-instance coordination.
type ExternalBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventType EventType, handler ExtEventHandler)
}
