package externalbus

import (
	"context"
	"encoding/json"
	"fmt"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/logger"
	"mizu/internal/domain/common"
	"sync"

	"github.com/redis/go-redis/v9"
)

const redisChannel = "mizu:events"

type envelope struct {
	Type common.EventType `json:"type"`
	Data json.RawMessage  `json:"data"`
}

// RedisBus is a Redis Pub/Sub implementation of ExternalBus.
//
// Events are serialized and published to a shared Redis channel.
// Concrete event types must be registered so received events can
// be reconstructed before dispatching to handlers.
type RedisBus struct {
	mu       sync.RWMutex
	client   *redis.Client
	handlers map[common.EventType][]eventbus.EventHandler

	// Used to reconstruct concrete event types when
	// receiving events from Redis.
	factories map[common.EventType]func() common.Event

	logger logger.Logger
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewRedisBus(
	client *redis.Client,
	logger logger.Logger,
) *RedisBus {
	return &RedisBus{
		client:    client,
		handlers:  make(map[common.EventType][]eventbus.EventHandler),
		factories: make(map[common.EventType]func() common.Event),
		logger:    logger,
	}
}

// RegisterEvent registers a factory used to reconstruct
// a concrete event when received from Redis.
func (b *RedisBus) RegisterEvent(
	eventType common.EventType,
	factory func() common.Event,
) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.factories[eventType] = factory
}

// Subscribe registers a handler for the given event type.
// Subscribe all handlers before calling Start.
func (b *RedisBus) Subscribe(
	eventType common.EventType,
	handler eventbus.EventHandler,
) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(
		b.handlers[eventType],
		handler,
	)
}

// Publish serializes and publishes an event to Redis.
func (b *RedisBus) Publish(
	ctx context.Context,
	event common.Event,
) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf(
			"externalbus.RedisBus.Publish: marshal event: %w",
			err,
		)
	}

	payload, err := json.Marshal(envelope{
		Type: event.EventType(),
		Data: data,
	})
	if err != nil {
		return fmt.Errorf(
			"externalbus.RedisBus.Publish: marshal envelope: %w",
			err,
		)
	}

	if err := b.client.Publish(
		ctx,
		redisChannel,
		payload,
	).Err(); err != nil {
		return fmt.Errorf(
			"externalbus.RedisBus.Publish: %w",
			err,
		)
	}

	return nil
}

// Start begins listening for events from Redis.
func (b *RedisBus) Start(ctx context.Context) {
	ctx, b.cancel = context.WithCancel(ctx)

	pubsub := b.client.Subscribe(ctx, redisChannel)

	if _, err := pubsub.Receive(ctx); err != nil {
		b.logger.Error(
			"externalbus.RedisBus: subscribe failed",
			"err", err,
		)
		return
	}

	b.wg.Go(func() {
		defer pubsub.Close()

		ch := pubsub.Channel()

		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}

				b.dispatch(ctx, msg.Payload)

			case <-ctx.Done():
				return
			}
		}
	})
}

// Stop signals the bus to stop and waits for the
// subscriber goroutine to exit.
func (b *RedisBus) Stop() {
	if b.cancel != nil {
		b.cancel()
	}

	b.wg.Wait()
}

func (b *RedisBus) dispatch(
	ctx context.Context,
	payload string,
) {
	var env envelope

	if err := json.Unmarshal(
		[]byte(payload),
		&env,
	); err != nil {
		b.logger.Error(
			"externalbus.RedisBus: failed to unmarshal envelope",
			"err", err,
		)
		return
	}

	b.mu.RLock()

	factory, ok := b.factories[env.Type]
	if !ok {
		b.mu.RUnlock()

		b.logger.Error(
			"externalbus.RedisBus: no factory registered",
			"event_type", env.Type,
		)

		return
	}

	event := factory()

	handlers := b.handlers[env.Type]

	b.mu.RUnlock()

	if err := json.Unmarshal(
		env.Data,
		event,
	); err != nil {
		b.logger.Error(
			"externalbus.RedisBus: failed to unmarshal event",
			"event_type", env.Type,
			"err", err,
		)
		return
	}

	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			b.logger.Error(
				"externalbus.RedisBus: handler error",
				"event_type", env.Type,
				"err", err,
			)
		}
	}
}
