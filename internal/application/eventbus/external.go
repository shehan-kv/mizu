package eventbus

import (
	"context"
	"mizu/internal/domain/common"
)

// ExternalBus delivers events across all running instances.
// Used for real-time notifications, SSE push, and cross-instance coordination.
type ExternalBus interface {
	Publish(ctx context.Context, event common.Event) error
	Subscribe(eventType common.EventType, handler EventHandler)
}
