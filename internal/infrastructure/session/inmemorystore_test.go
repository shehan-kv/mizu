package session

import (
	"context"
	"errors"
	"testing"

	appsession "mizu/internal/application/session"
	"mizu/internal/domain/iam"
)

func TestNewInMemoryStore(t *testing.T) {
	store := NewInMemoryStore()

	if store == nil {
		t.Fatal("expected store to be non-nil")
	}

	if store.sessions == nil {
		t.Fatal("expected sessions map to be initialized")
	}

	if store.userSessionIDs == nil {
		t.Fatal("expected userSessionIDs map to be initialized")
	}
}

func TestInMemoryStore_Add(t *testing.T) {
	t.Run("stores session", func(t *testing.T) {
		store := NewInMemoryStore()
		sess := newTestSession(t, "session-1", "user-1")

		err := store.Add(context.Background(), sess)
		if err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		got, err := store.Get(context.Background(), sess.ID())
		if err != nil {
			t.Fatalf("Get() returned error: %v", err)
		}

		if got != sess {
			t.Fatal("expected stored session to be the same session")
		}
	})

	t.Run("associates session with user", func(t *testing.T) {
		store := NewInMemoryStore()
		sess := newTestSession(t, "session-1", "user-1")

		if err := store.Add(context.Background(), sess); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		sessions, err := store.ListByUser(
			context.Background(),
			sess.UserID(),
		)
		if err != nil {
			t.Fatalf("ListByUser() returned error: %v", err)
		}

		if len(sessions) != 1 {
			t.Fatalf("expected 1 session, got %d", len(sessions))
		}

		if sessions[0] != sess {
			t.Fatal("expected listed session to be the added session")
		}
	})
}

func TestInMemoryStore_Get(t *testing.T) {
	t.Run("returns existing session", func(t *testing.T) {
		store := NewInMemoryStore()
		sess := newTestSession(t, "session-1", "user-1")

		if err := store.Add(context.Background(), sess); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		got, err := store.Get(context.Background(), sess.ID())
		if err != nil {
			t.Fatalf("Get() returned error: %v", err)
		}

		if got != sess {
			t.Fatal("expected Get() to return the stored session")
		}
	})

	t.Run("returns ErrSessionNotFound for missing session", func(t *testing.T) {
		store := NewInMemoryStore()

		id, err := appsession.NewSessionID("missing-session")
		if err != nil {
			t.Fatalf("failed to create session ID: %v", err)
		}

		got, err := store.Get(context.Background(), id)

		if got != nil {
			t.Fatal("expected nil session")
		}

		if !errors.Is(err, appsession.ErrSessionNotFound) {
			t.Fatalf(
				"expected ErrSessionNotFound, got %v",
				err,
			)
		}
	})
}

func TestInMemoryStore_ListByUser(t *testing.T) {
	t.Run("returns nil when user has no sessions", func(t *testing.T) {
		store := NewInMemoryStore()

		userID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatalf("failed to create user ID: %v", err)
		}

		sessions, err := store.ListByUser(
			context.Background(),
			userID,
		)
		if err != nil {
			t.Fatalf("ListByUser() returned error: %v", err)
		}

		if sessions != nil {
			t.Fatalf("expected nil, got %v", sessions)
		}
	})

	t.Run("returns all sessions for user", func(t *testing.T) {
		store := NewInMemoryStore()

		first := newTestSession(t, "session-1", "user-1")
		second := newTestSession(t, "session-2", "user-1")

		if err := store.Add(context.Background(), first); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		if err := store.Add(context.Background(), second); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		sessions, err := store.ListByUser(
			context.Background(),
			first.UserID(),
		)
		if err != nil {
			t.Fatalf("ListByUser() returned error: %v", err)
		}

		if len(sessions) != 2 {
			t.Fatalf("expected 2 sessions, got %d", len(sessions))
		}

		if sessions[0] != first {
			t.Fatal("expected first session at index 0")
		}

		if sessions[1] != second {
			t.Fatal("expected second session at index 1")
		}
	})

	t.Run("returns only sessions belonging to requested user", func(t *testing.T) {
		store := NewInMemoryStore()

		userOneSession := newTestSession(t, "session-1", "user-1")
		userTwoSession := newTestSession(t, "session-2", "user-2")

		if err := store.Add(context.Background(), userOneSession); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		if err := store.Add(context.Background(), userTwoSession); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		sessions, err := store.ListByUser(
			context.Background(),
			userOneSession.UserID(),
		)
		if err != nil {
			t.Fatalf("ListByUser() returned error: %v", err)
		}

		if len(sessions) != 1 {
			t.Fatalf("expected 1 session, got %d", len(sessions))
		}

		if sessions[0] != userOneSession {
			t.Fatal("returned session belongs to the wrong user")
		}
	})
}

func TestInMemoryStore_Delete(t *testing.T) {
	t.Run("deletes existing session", func(t *testing.T) {
		store := NewInMemoryStore()
		sess := newTestSession(t, "session-1", "user-1")

		if err := store.Add(context.Background(), sess); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		if err := store.Delete(context.Background(), sess.ID()); err != nil {
			t.Fatalf("Delete() returned error: %v", err)
		}

		_, err := store.Get(context.Background(), sess.ID())
		if !errors.Is(err, appsession.ErrSessionNotFound) {
			t.Fatalf(
				"expected ErrSessionNotFound after deletion, got %v",
				err,
			)
		}
	})

	t.Run("removes session from user's session list", func(t *testing.T) {
		store := NewInMemoryStore()

		first := newTestSession(t, "session-1", "user-1")
		second := newTestSession(t, "session-2", "user-1")

		if err := store.Add(context.Background(), first); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		if err := store.Add(context.Background(), second); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		if err := store.Delete(context.Background(), first.ID()); err != nil {
			t.Fatalf("Delete() returned error: %v", err)
		}

		sessions, err := store.ListByUser(
			context.Background(),
			first.UserID(),
		)
		if err != nil {
			t.Fatalf("ListByUser() returned error: %v", err)
		}

		if len(sessions) != 1 {
			t.Fatalf("expected 1 session, got %d", len(sessions))
		}

		if sessions[0] != second {
			t.Fatal("expected remaining session to be second")
		}
	})

	t.Run("removes user entry when last session is deleted", func(t *testing.T) {
		store := NewInMemoryStore()
		sess := newTestSession(t, "session-1", "user-1")

		if err := store.Add(context.Background(), sess); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		if err := store.Delete(context.Background(), sess.ID()); err != nil {
			t.Fatalf("Delete() returned error: %v", err)
		}

		sessions, err := store.ListByUser(
			context.Background(),
			sess.UserID(),
		)
		if err != nil {
			t.Fatalf("ListByUser() returned error: %v", err)
		}

		if sessions != nil {
			t.Fatalf("expected nil sessions, got %v", sessions)
		}

		if _, ok := store.userSessionIDs[sess.UserID()]; ok {
			t.Fatal("expected user session entry to be removed")
		}
	})

	t.Run("succeeds when session does not exist", func(t *testing.T) {
		store := NewInMemoryStore()

		id, err := appsession.NewSessionID("missing-session")
		if err != nil {
			t.Fatalf("failed to create session ID: %v", err)
		}

		if err := store.Delete(context.Background(), id); err != nil {
			t.Fatalf("Delete() returned error: %v", err)
		}
	})

	t.Run("can delete one session while keeping other users' sessions", func(t *testing.T) {
		store := NewInMemoryStore()

		userOne := newTestSession(t, "session-1", "user-1")
		userTwo := newTestSession(t, "session-2", "user-2")

		if err := store.Add(context.Background(), userOne); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		if err := store.Add(context.Background(), userTwo); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		if err := store.Delete(context.Background(), userOne.ID()); err != nil {
			t.Fatalf("Delete() returned error: %v", err)
		}

		sessions, err := store.ListByUser(
			context.Background(),
			userTwo.UserID(),
		)
		if err != nil {
			t.Fatalf("ListByUser() returned error: %v", err)
		}

		if len(sessions) != 1 {
			t.Fatalf("expected 1 session for user 2, got %d", len(sessions))
		}

		if sessions[0] != userTwo {
			t.Fatal("user 2's session was affected by deleting user 1's session")
		}
	})
}
