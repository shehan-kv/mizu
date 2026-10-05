package internalbus

import (
	"context"
	"errors"
	"mizu/internal/application/logger"
	"mizu/internal/domain/common"
	"testing"
	"time"
)

type testEvent struct {
	eventType common.EventType
}

func (e testEvent) EventType() common.EventType {
	return e.eventType
}

type fakeEventHandler struct {
	calls int
	event common.Event
	err   error
}

func (f *fakeEventHandler) handle(
	_ context.Context,
	event common.Event,
) error {
	f.calls++
	f.event = event
	return f.err
}

type fakeLogger struct {
	warnCalls  int
	errorCalls int
}

func (f *fakeLogger) Info(string, ...any) {}

func (f *fakeLogger) Warn(string, ...any) {
	f.warnCalls++
}

func (f *fakeLogger) Error(string, ...any) {
	f.errorCalls++
}

func (f *fakeLogger) Fatal(string, ...any) {}

var _ logger.Logger = (*fakeLogger)(nil)

func waitForCondition(
	t *testing.T,
	condition func() bool,
) {
	t.Helper()

	deadline := time.Now().Add(time.Second)

	for time.Now().Before(deadline) {
		if condition() {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("condition was not met before timeout")
}

func TestInMemoryBus_Publish(t *testing.T) {
	t.Run("dispatches event to subscribed handler", func(t *testing.T) {
		log := &fakeLogger{}
		bus := NewInMemoryBus(log, 1)

		eventType := common.EventType("test.event")
		event := testEvent{
			eventType: eventType,
		}

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
		defer bus.Stop()

		err := bus.Publish(ctx, event)
		if err != nil {
			t.Fatalf(
				"Publish() error = %v",
				err,
			)
		}

		waitForCondition(t, func() bool {
			return handler.calls == 1
		})

		if handler.event != event {
			t.Errorf(
				"handler event = %#v, want %#v",
				handler.event,
				event,
			)
		}
	})
}

func TestInMemoryBus_MultipleHandlers(t *testing.T) {
	t.Run("dispatches event to all subscribed handlers", func(t *testing.T) {
		log := &fakeLogger{}
		bus := NewInMemoryBus(log, 1)

		eventType := common.EventType("test.event")
		event := testEvent{
			eventType: eventType,
		}

		first := &fakeEventHandler{}
		second := &fakeEventHandler{}

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

		waitForCondition(t, func() bool {
			return first.calls == 1 &&
				second.calls == 1
		})
	})
}

func TestInMemoryBus_HandlerError(t *testing.T) {
	t.Run("continues dispatching after handler error", func(t *testing.T) {
		log := &fakeLogger{}
		bus := NewInMemoryBus(log, 1)

		eventType := common.EventType("test.event")
		event := testEvent{
			eventType: eventType,
		}

		handlerErr := errors.New("handler failed")

		first := &fakeEventHandler{
			err: handlerErr,
		}

		second := &fakeEventHandler{}

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

		waitForCondition(t, func() bool {
			return second.calls == 1
		})

		if first.calls != 1 {
			t.Errorf(
				"first handler calls = %d, want 1",
				first.calls,
			)
		}

		if second.calls != 1 {
			t.Errorf(
				"second handler calls = %d, want 1",
				second.calls,
			)
		}

		if log.errorCalls != 1 {
			t.Errorf(
				"logger error calls = %d, want 1",
				log.errorCalls,
			)
		}
	})
}

func TestInMemoryBus_EventType(t *testing.T) {
	t.Run("dispatches only to matching handlers", func(t *testing.T) {
		log := &fakeLogger{}
		bus := NewInMemoryBus(log, 2)

		firstType := common.EventType("first.event")
		secondType := common.EventType("second.event")

		firstEvent := testEvent{
			eventType: firstType,
		}

		secondEvent := testEvent{
			eventType: secondType,
		}

		firstHandler := &fakeEventHandler{}
		secondHandler := &fakeEventHandler{}

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

		waitForCondition(t, func() bool {
			return firstHandler.calls == 1 &&
				secondHandler.calls == 1
		})

		if firstHandler.event != firstEvent {
			t.Errorf(
				"first handler event = %#v, want %#v",
				firstHandler.event,
				firstEvent,
			)
		}

		if secondHandler.event != secondEvent {
			t.Errorf(
				"second handler event = %#v, want %#v",
				secondHandler.event,
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
			eventType: common.EventType("test.event"),
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

		if log.warnCalls != 1 {
			t.Errorf(
				"logger warn calls = %d, want 1",
				log.warnCalls,
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
				eventType: common.EventType("test.event"),
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

		eventType := common.EventType("test.event")

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

		if handler.calls != 2 {
			t.Errorf(
				"handler calls = %d, want 2",
				handler.calls,
			)
		}
	})
}
