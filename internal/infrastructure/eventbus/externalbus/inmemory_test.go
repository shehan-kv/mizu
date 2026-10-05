package externalbus

import (
	"context"
	"errors"
	"mizu/internal/application/eventbus"
	"testing"
)

func TestInMemoryBus_Publish(t *testing.T) {
	t.Run("dispatches event to subscribed handler", func(t *testing.T) {
		log := &fakeLogger{}
		bus := NewInMemoryBus(log, 1)

		eventType := eventbus.EventType("test.event")
		event := testEvent{
			eventType: eventType,
		}

		handler := &fakeEventHandler{
			done: make(chan struct{}),
		}

		bus.Subscribe(
			eventType,
			handler.handle,
		)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)
		defer cancel()

		bus.Start(ctx)
		defer bus.Stop()

		err := bus.Publish(ctx, event)
		if err != nil {
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

		if handler.Event() != event {
			t.Errorf(
				"handler event = %#v, want %#v",
				handler.Event(),
				event,
			)
		}
	})
}

func TestInMemoryBus_MultipleHandlers(t *testing.T) {
	t.Run("dispatches event to all subscribed handlers", func(t *testing.T) {
		log := &fakeLogger{}
		bus := NewInMemoryBus(log, 1)

		eventType := eventbus.EventType("test.event")
		event := testEvent{
			eventType: eventType,
		}

		first := &fakeEventHandler{
			done: make(chan struct{}),
		}

		second := &fakeEventHandler{
			done: make(chan struct{}),
		}

		bus.Subscribe(
			eventType,
			first.handle,
		)

		bus.Subscribe(
			eventType,
			second.handle,
		)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)
		defer cancel()

		bus.Start(ctx)
		defer bus.Stop()

		err := bus.Publish(ctx, event)
		if err != nil {
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
}

func TestInMemoryBus_HandlerError(t *testing.T) {
	t.Run("continues dispatching after handler error", func(t *testing.T) {
		log := &fakeLogger{}
		bus := NewInMemoryBus(log, 1)

		eventType := eventbus.EventType("test.event")
		event := testEvent{
			eventType: eventType,
		}

		handlerErr := errors.New("handler failed")

		first := &fakeEventHandler{
			err:  handlerErr,
			done: make(chan struct{}),
		}

		second := &fakeEventHandler{
			done: make(chan struct{}),
		}

		bus.Subscribe(
			eventType,
			first.handle,
		)

		bus.Subscribe(
			eventType,
			second.handle,
		)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)
		defer cancel()

		bus.Start(ctx)
		defer bus.Stop()

		err := bus.Publish(ctx, event)
		if err != nil {
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
}

func TestInMemoryBus_EventType(t *testing.T) {
	t.Run("dispatches only to matching handlers", func(t *testing.T) {
		log := &fakeLogger{}
		bus := NewInMemoryBus(log, 2)

		firstType := eventbus.EventType("first.event")
		secondType := eventbus.EventType("second.event")

		firstEvent := testEvent{
			eventType: firstType,
		}

		secondEvent := testEvent{
			eventType: secondType,
		}

		firstHandler := &fakeEventHandler{
			done: make(chan struct{}),
		}

		secondHandler := &fakeEventHandler{
			done: make(chan struct{}),
		}

		bus.Subscribe(
			firstType,
			firstHandler.handle,
		)

		bus.Subscribe(
			secondType,
			secondHandler.handle,
		)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)
		defer cancel()

		bus.Start(ctx)
		defer bus.Stop()

		if err := bus.Publish(ctx, firstEvent); err != nil {
			t.Fatalf(
				"Publish(firstEvent) error = %v",
				err,
			)
		}

		if err := bus.Publish(ctx, secondEvent); err != nil {
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

		if firstHandler.Event() != firstEvent {
			t.Errorf(
				"first handler event = %#v, want %#v",
				firstHandler.Event(),
				firstEvent,
			)
		}

		if secondHandler.Event() != secondEvent {
			t.Errorf(
				"second handler event = %#v, want %#v",
				secondHandler.Event(),
				secondEvent,
			)
		}
	})
}

func TestInMemoryBus_PublishFull(t *testing.T) {
	t.Run("drops event when buffer is full", func(t *testing.T) {
		log := &fakeLogger{}
		bus := NewInMemoryBus(log, 1)

		event := testEvent{
			eventType: eventbus.EventType("test.event"),
		}

		if err := bus.Publish(
			context.Background(),
			event,
		); err != nil {
			t.Fatalf(
				"first Publish() error = %v",
				err,
			)
		}

		if err := bus.Publish(
			context.Background(),
			event,
		); err != nil {
			t.Fatalf(
				"second Publish() error = %v",
				err,
			)
		}

		if log.WarnCalls() != 1 {
			t.Errorf(
				"logger warn calls = %d, want 1",
				log.WarnCalls(),
			)
		}
	})
}

func TestInMemoryBus_PublishCancelled(t *testing.T) {
	t.Run("returns context error when context is cancelled", func(t *testing.T) {
		log := &fakeLogger{}
		bus := NewInMemoryBus(log, 1)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)
		cancel()

		err := bus.Publish(
			ctx,
			testEvent{
				eventType: eventbus.EventType("test.event"),
			},
		)

		if !errors.Is(err, context.Canceled) {
			t.Fatalf(
				"Publish() error = %v, want %v",
				err,
				context.Canceled,
			)
		}
	})
}

func TestInMemoryBus_Stop(t *testing.T) {
	t.Run("drains queued events before returning", func(t *testing.T) {
		log := &fakeLogger{}
		bus := NewInMemoryBus(log, 2)

		eventType := eventbus.EventType("test.event")

		handler := &fakeEventHandler{}

		bus.Subscribe(
			eventType,
			handler.handle,
		)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)
		defer cancel()

		bus.Start(ctx)

		for range 2 {
			if err := bus.Publish(
				context.Background(),
				testEvent{
					eventType: eventType,
				},
			); err != nil {
				t.Fatalf(
					"Publish() error = %v",
					err,
				)
			}
		}

		bus.Stop()

		if handler.Calls() != 2 {
			t.Errorf(
				"handler calls = %d, want 2",
				handler.Calls(),
			)
		}
	})
}
