package externalbus

import (
	"context"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/logger"
	"sync"
)

// InMemoryBus is the in-memory implementation of port.ExternalBus.
// Use for single-instance deployments.
// Replace with Redis Pub/Sub or RabbitMQ for multi-instance.
type InMemoryBus struct {
	mu       sync.RWMutex
	handlers map[eventbus.EventType][]eventbus.ExtEventHandler
	ch       chan eventbus.Event
	logger   logger.Logger
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewInMemoryBus(logger logger.Logger, buffer int) *InMemoryBus {
	return &InMemoryBus{
		handlers: make(map[eventbus.EventType][]eventbus.ExtEventHandler),
		ch:       make(chan eventbus.Event, buffer),
		logger:   logger,
	}
}

func (b *InMemoryBus) Subscribe(eventType eventbus.EventType, handler eventbus.ExtEventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *InMemoryBus) Publish(ctx context.Context, event eventbus.Event) error {
	select {
	case b.ch <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		b.logger.Warn("external bus full, dropping event",
			"event_type", event.EventType(),
		)
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

// Stop signals the bus to stop and waits for all in-flight events to finish.
func (b *InMemoryBus) Stop() {
	b.cancel()
	b.wg.Wait()
}

func (b *InMemoryBus) dispatch(ctx context.Context, event eventbus.Event) {
	b.mu.RLock()
	handlers := b.handlers[event.EventType()]
	b.mu.RUnlock()
	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			b.logger.Error("external bus handler error",
				"event_type", event.EventType(),
				"err", err,
			)
		}
	}
}
