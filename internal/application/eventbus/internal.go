package eventbus

import (
	"context"
	"mizu/internal/domain/common"
)

type EventHandler func(ctx context.Context, event common.Event) error

// InternalBus delivers events within the current process.
// Events are guaranteed to be handled by the same instance that published them.
type InternalBus interface {
	Publish(ctx context.Context, event common.Event) error
	Subscribe(eventType common.EventType, handler EventHandler)
}
