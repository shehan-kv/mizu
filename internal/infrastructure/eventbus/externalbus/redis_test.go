package externalbus

import (
	"context"
	"encoding/json"
	"errors"
	"mizu/internal/application/eventbus"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func newTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()

	ctx := context.Background()

	container, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "redis:8.2-alpine",
				ExposedPorts: []string{"6379/tcp"},
				WaitingFor:   wait.ForListeningPort("6379/tcp"),
			},
			Started: true,
		},
	)
	if err != nil {
		t.Fatalf(
			"failed to start Redis container: %v",
			err,
		)
	}

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf(
			"failed to get Redis container host: %v",
			err,
		)
	}

	port, err := container.MappedPort(
		ctx,
		"6379/tcp",
	)
	if err != nil {
		t.Fatalf(
			"failed to get Redis container port: %v",
			err,
		)
	}

	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})

	t.Cleanup(func() {
		_ = client.Close()
	})

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf(
			"failed to connect to Redis: %v",
			err,
		)
	}

	return client
}

func resetRedis(
	t *testing.T,
	client *redis.Client,
) {
	t.Helper()

	if err := client.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf(
			"failed to flush Redis database: %v",
			err,
		)
	}
}

func waitForLogError(
	t *testing.T,
	log *fakeLogger,
) {
	t.Helper()

	deadline := time.Now().Add(time.Second)

	for time.Now().Before(deadline) {
		if log.ErrorCalls() == 1 {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("expected logger error call before timeout")
}

type redisTestEvent struct {
	Value string `json:"value"`
}

func (e redisTestEvent) EventType() eventbus.EventType {
	return eventbus.EventType("redis.test.event")
}

func TestRedisBus(t *testing.T) {
	client := newTestRedisClient(t)

	t.Run("publishes event to Redis", func(t *testing.T) {
		resetRedis(t, client)

		ctx := context.Background()
		log := &fakeLogger{}

		bus := NewRedisBus(
			client,
			log,
		)

		event := testEvent{
			eventType: eventbus.EventType("test.event"),
		}

		pubsub := client.Subscribe(
			ctx,
			redisChannel,
		)
		defer pubsub.Close()

		if _, err := pubsub.Receive(ctx); err != nil {
			t.Fatalf(
				"failed to subscribe: %v",
				err,
			)
		}

		if err := bus.Publish(ctx, event); err != nil {
			t.Fatalf(
				"Publish() error = %v",
				err,
			)
		}

		select {
		case msg := <-pubsub.Channel():
			if msg.Channel != redisChannel {
				t.Errorf(
					"message channel = %q, want %q",
					msg.Channel,
					redisChannel,
				)
			}

			var env envelope

			if err := json.Unmarshal(
				[]byte(msg.Payload),
				&env,
			); err != nil {
				t.Fatalf(
					"failed to unmarshal published envelope: %v",
					err,
				)
			}

			if env.Type != event.EventType() {
				t.Errorf(
					"envelope type = %q, want %q",
					env.Type,
					event.EventType(),
				)
			}

			var received testEvent

			if err := json.Unmarshal(
				env.Data,
				&received,
			); err != nil {
				t.Fatalf(
					"failed to unmarshal event data: %v",
					err,
				)
			}

		case <-time.After(time.Second):
			t.Fatal("did not receive published event")
		}
	})

	t.Run("dispatches published event to subscribed handler", func(t *testing.T) {
		resetRedis(t, client)

		ctx := context.Background()
		log := &fakeLogger{}

		bus := NewRedisBus(
			client,
			log,
		)

		eventType := eventbus.EventType("test.event")

		handler := &fakeEventHandler{
			done: make(chan struct{}),
		}

		bus.RegisterEvent(
			eventType,
			func() eventbus.Event {
				return &testEvent{
					eventType: eventType,
				}
			},
		)

		bus.Subscribe(
			eventType,
			handler.handle,
		)

		bus.Start(ctx)
		defer bus.Stop()

		event := testEvent{
			eventType: eventType,
		}

		if err := bus.Publish(ctx, event); err != nil {
			t.Fatalf(
				"Publish() error = %v",
				err,
			)
		}

		waitForHandler(t, handler)

		if handler.Calls() != 1 {
			t.Errorf(
				"handler calls = %d, want 1",
				handler.Calls(),
			)
		}

		received, ok := handler.Event().(*testEvent)
		if !ok {
			t.Fatalf(
				"handler event type = %T, want *testEvent",
				handler.Event(),
			)
		}

		if received.EventType() != eventType {
			t.Errorf(
				"handler event type = %q, want %q",
				received.EventType(),
				eventType,
			)
		}
	})

	t.Run("dispatches event to all subscribed handlers", func(t *testing.T) {
		resetRedis(t, client)

		ctx := context.Background()
		log := &fakeLogger{}

		bus := NewRedisBus(
			client,
			log,
		)

		eventType := eventbus.EventType("test.event")

		first := &fakeEventHandler{
			done: make(chan struct{}),
		}

		second := &fakeEventHandler{
			done: make(chan struct{}),
		}

		bus.RegisterEvent(
			eventType,
			func() eventbus.Event {
				return &testEvent{
					eventType: eventType,
				}
			},
		)

		bus.Subscribe(
			eventType,
			first.handle,
		)

		bus.Subscribe(
			eventType,
			second.handle,
		)

		bus.Start(ctx)
		defer bus.Stop()

		event := testEvent{
			eventType: eventType,
		}

		if err := bus.Publish(ctx, event); err != nil {
			t.Fatalf(
				"Publish() error = %v",
				err,
			)
		}

		waitForHandler(t, first)
		waitForHandler(t, second)

		if first.Calls() != 1 {
			t.Errorf(
				"first handler calls = %d, want 1",
				first.Calls(),
			)
		}

		if second.Calls() != 1 {
			t.Errorf(
				"second handler calls = %d, want 1",
				second.Calls(),
			)
		}
	})

	t.Run("continues dispatching after handler error", func(t *testing.T) {
		resetRedis(t, client)

		ctx := context.Background()
		log := &fakeLogger{}

		bus := NewRedisBus(
			client,
			log,
		)

		eventType := eventbus.EventType("test.event")
		handlerErr := errors.New("handler failed")

		first := &fakeEventHandler{
			err:  handlerErr,
			done: make(chan struct{}),
		}

		second := &fakeEventHandler{
			done: make(chan struct{}),
		}

		bus.RegisterEvent(
			eventType,
			func() eventbus.Event {
				return &testEvent{
					eventType: eventType,
				}
			},
		)

		bus.Subscribe(
			eventType,
			first.handle,
		)

		bus.Subscribe(
			eventType,
			second.handle,
		)

		bus.Start(ctx)
		defer bus.Stop()

		if err := bus.Publish(
			ctx,
			testEvent{
				eventType: eventType,
			},
		); err != nil {
			t.Fatalf(
				"Publish() error = %v",
				err,
			)
		}

		waitForHandler(t, second)

		if first.Calls() != 1 {
			t.Errorf(
				"first handler calls = %d, want 1",
				first.Calls(),
			)
		}

		if second.Calls() != 1 {
			t.Errorf(
				"second handler calls = %d, want 1",
				second.Calls(),
			)
		}

		if log.ErrorCalls() != 1 {
			t.Errorf(
				"logger error calls = %d, want 1",
				log.ErrorCalls(),
			)
		}
	})

	t.Run("dispatches only to matching handlers", func(t *testing.T) {
		resetRedis(t, client)

		ctx := context.Background()
		log := &fakeLogger{}

		bus := NewRedisBus(
			client,
			log,
		)

		firstType := eventbus.EventType("first.event")
		secondType := eventbus.EventType("second.event")

		firstHandler := &fakeEventHandler{
			done: make(chan struct{}),
		}

		secondHandler := &fakeEventHandler{
			done: make(chan struct{}),
		}

		bus.RegisterEvent(
			firstType,
			func() eventbus.Event {
				return &testEvent{
					eventType: firstType,
				}
			},
		)

		bus.RegisterEvent(
			secondType,
			func() eventbus.Event {
				return &testEvent{
					eventType: secondType,
				}
			},
		)

		bus.Subscribe(
			firstType,
			firstHandler.handle,
		)

		bus.Subscribe(
			secondType,
			secondHandler.handle,
		)

		bus.Start(ctx)
		defer bus.Stop()

		if err := bus.Publish(
			ctx,
			testEvent{
				eventType: firstType,
			},
		); err != nil {
			t.Fatalf(
				"Publish(firstEvent) error = %v",
				err,
			)
		}

		if err := bus.Publish(
			ctx,
			testEvent{
				eventType: secondType,
			},
		); err != nil {
			t.Fatalf(
				"Publish(secondEvent) error = %v",
				err,
			)
		}

		waitForHandler(t, firstHandler)
		waitForHandler(t, secondHandler)

		if firstHandler.Calls() != 1 {
			t.Errorf(
				"first handler calls = %d, want 1",
				firstHandler.Calls(),
			)
		}

		if secondHandler.Calls() != 1 {
			t.Errorf(
				"second handler calls = %d, want 1",
				secondHandler.Calls(),
			)
		}

		if firstHandler.Event().EventType() != firstType {
			t.Errorf(
				"first handler event type = %q, want %q",
				firstHandler.Event().EventType(),
				firstType,
			)
		}

		if secondHandler.Event().EventType() != secondType {
			t.Errorf(
				"second handler event type = %q, want %q",
				secondHandler.Event().EventType(),
				secondType,
			)
		}
	})

	t.Run("logs error when event factory is not registered", func(t *testing.T) {
		resetRedis(t, client)

		ctx := context.Background()
		log := &fakeLogger{}

		bus := NewRedisBus(
			client,
			log,
		)

		bus.Start(ctx)
		defer bus.Stop()

		if err := bus.Publish(
			ctx,
			testEvent{
				eventType: eventbus.EventType("test.event"),
			},
		); err != nil {
			t.Fatalf(
				"Publish() error = %v",
				err,
			)
		}

		waitForLogError(t, log)
	})

	t.Run("logs error for invalid envelope", func(t *testing.T) {
		resetRedis(t, client)

		ctx := context.Background()
		log := &fakeLogger{}

		bus := NewRedisBus(
			client,
			log,
		)

		bus.Start(ctx)
		defer bus.Stop()

		if err := client.Publish(
			ctx,
			redisChannel,
			"invalid json",
		).Err(); err != nil {
			t.Fatalf(
				"failed to publish invalid envelope: %v",
				err,
			)
		}

		waitForLogError(t, log)
	})

	t.Run("logs error when event data cannot be unmarshaled", func(t *testing.T) {
		resetRedis(t, client)

		ctx := context.Background()
		log := &fakeLogger{}

		bus := NewRedisBus(
			client,
			log,
		)

		eventType := eventbus.EventType("redis.test.event")

		bus.RegisterEvent(
			eventType,
			func() eventbus.Event {
				return &redisTestEvent{}
			},
		)

		bus.Start(ctx)
		defer bus.Stop()

		payload := `{
			"type": "redis.test.event",
			"data": {
				"value": 123
			}
		}`

		if err := client.Publish(
			ctx,
			redisChannel,
			payload,
		).Err(); err != nil {
			t.Fatalf(
				"failed to publish invalid event: %v",
				err,
			)
		}

		waitForLogError(t, log)
	})

	t.Run("returns context error when context is cancelled", func(t *testing.T) {
		resetRedis(t, client)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)
		cancel()

		log := &fakeLogger{}

		bus := NewRedisBus(
			client,
			log,
		)

		err := bus.Publish(
			ctx,
			testEvent{
				eventType: eventbus.EventType("test.event"),
			},
		)

		if err == nil {
			t.Fatal(
				"Publish() error = nil, want context canceled",
			)
		}

		if !errors.Is(err, context.Canceled) {
			t.Fatalf(
				"Publish() error = %v, want context canceled",
				err,
			)
		}
	})

	t.Run("stops subscriber", func(t *testing.T) {
		resetRedis(t, client)

		ctx := context.Background()
		log := &fakeLogger{}

		bus := NewRedisBus(
			client,
			log,
		)

		eventType := eventbus.EventType("test.event")

		handler := &fakeEventHandler{
			done: make(chan struct{}),
		}

		bus.RegisterEvent(
			eventType,
			func() eventbus.Event {
				return &testEvent{
					eventType: eventType,
				}
			},
		)

		bus.Subscribe(
			eventType,
			handler.handle,
		)

		bus.Start(ctx)

		if err := bus.Publish(
			ctx,
			testEvent{
				eventType: eventType,
			},
		); err != nil {
			t.Fatalf(
				"Publish() error = %v",
				err,
			)
		}

		waitForHandler(t, handler)

		bus.Stop()

		if handler.Calls() != 1 {
			t.Errorf(
				"handler calls = %d, want 1",
				handler.Calls(),
			)
		}
	})
}
