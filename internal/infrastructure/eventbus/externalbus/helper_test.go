package externalbus

import (
	"context"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/logger"
	"sync"
	"testing"
	"time"
)

type testEvent struct {
	eventType eventbus.EventType
}

func (e testEvent) EventType() eventbus.EventType {
	return e.eventType
}

type fakeEventHandler struct {
	mu    sync.Mutex
	calls int
	event eventbus.Event
	err   error
	done  chan struct{}
}

func (f *fakeEventHandler) handle(
	_ context.Context,
	event eventbus.Event,
) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls++
	f.event = event

	if f.done != nil && f.calls == 1 {
		close(f.done)
	}

	return f.err
}

func (f *fakeEventHandler) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.calls
}

func (f *fakeEventHandler) Event() eventbus.Event {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.event
}

type fakeLogger struct {
	mu         sync.Mutex
	warnCalls  int
	errorCalls int
}

func (f *fakeLogger) Info(string, ...any) {}

func (f *fakeLogger) Warn(string, ...any) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.warnCalls++
}

func (f *fakeLogger) Error(string, ...any) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.errorCalls++
}

func (f *fakeLogger) Fatal(string, ...any) {}

func (f *fakeLogger) WarnCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.warnCalls
}

func (f *fakeLogger) ErrorCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.errorCalls
}

var _ logger.Logger = (*fakeLogger)(nil)

func waitForHandler(
	t *testing.T,
	handler *fakeEventHandler,
) {
	t.Helper()

	select {
	case <-handler.done:
	case <-time.After(time.Second):
		t.Fatal("handler was not called before timeout")
	}
}
