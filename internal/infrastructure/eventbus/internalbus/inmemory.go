package internalbus

import (
	"context"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/logger"
	"mizu/internal/domain/common"
	"sync"
)

type InMemoryBus struct {
	mu       sync.RWMutex
	handlers map[common.EventType][]eventbus.EventHandler
	ch       chan common.Event
	logger   logger.Logger
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewInMemoryBus(logger logger.Logger, buffer int) *InMemoryBus {
	return &InMemoryBus{
		handlers: make(map[common.EventType][]eventbus.EventHandler),
		ch:       make(chan common.Event, buffer),
		logger:   logger,
	}
}

func (b *InMemoryBus) Subscribe(eventType common.EventType, handler eventbus.EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *InMemoryBus) Publish(ctx context.Context, event common.Event) error {
	select {
	case b.ch <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		b.logger.Warn("internal bus full, dropping event", "event_type", event.EventType())
		return nil
	}
}

// Start begins processing events in a background goroutine.
// Subscribe all handlers before calling Start to avoid missing events.
func (b *InMemoryBus) Start(ctx context.Context) {
	ctx, b.cancel = context.WithCancel(ctx)
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		for {
			select {
			case event := <-b.ch:
				b.dispatch(ctx, event)
			case <-ctx.Done():
				for {
					select {
					case event := <-b.ch:
						b.dispatch(ctx, event)
					default:
						return
					}
				}
			}
		}
	}()
}

// Stop signals the bus to stop and waits for all in-flight events to be processed.
func (b *InMemoryBus) Stop() {
	b.cancel()
	b.wg.Wait()
}

func (b *InMemoryBus) dispatch(ctx context.Context, event common.Event) {
	b.mu.RLock()
	handlers := b.handlers[event.EventType()]
	b.mu.RUnlock()
	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			b.logger.Error("internal bus handler error",
				"event_type", event.EventType(),
				"err", err,
			)
		}
	}
}
