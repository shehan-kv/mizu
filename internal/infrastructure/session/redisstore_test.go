package session

import (
	"context"
	"errors"
	"testing"
	"time"

	appsession "mizu/internal/application/session"
	"mizu/internal/domain/iam"

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
		t.Fatalf("failed to start Redis container: %v", err)
	}

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get Redis host: %v", err)
	}

	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		t.Fatalf("failed to get Redis port: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})

	t.Cleanup(func() {
		_ = client.Close()
	})

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("failed to ping Redis: %v", err)
	}

	return client
}

func resetRedis(t *testing.T, client *redis.Client) {
	t.Helper()

	if err := client.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("failed to flush Redis database: %v", err)
	}
}

func TestRedisStore(t *testing.T) {
	client := newTestRedisClient(t)

	t.Run("adds and gets session", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, time.Hour)
		sess := newTestSession(t, "session-1", "user-1")

		err := store.Add(context.Background(), sess)
		if err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		got, err := store.Get(context.Background(), sess.ID())
		if err != nil {
			t.Fatalf("Get() returned error: %v", err)
		}

		if got.ID() != sess.ID() {
			t.Fatalf("expected session ID %q, got %q",
				sess.ID(),
				got.ID(),
			)
		}

		if got.UserID() != sess.UserID() {
			t.Fatalf("expected user ID %q, got %q",
				sess.UserID(),
				got.UserID(),
			)
		}
	})

	t.Run("stores session with expiry", func(t *testing.T) {
		resetRedis(t, client)

		duration := time.Hour
		store := NewRedisStore(client, duration)
		sess := newTestSession(t, "session-1", "user-1")

		if err := store.Add(context.Background(), sess); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		ttl, err := client.TTL(
			context.Background(),
			sessionKey(sess.ID()),
		).Result()
		if err != nil {
			t.Fatalf("failed to get session TTL: %v", err)
		}

		if ttl <= 0 || ttl > duration {
			t.Fatalf("expected TTL between 0 and %v, got %v", duration, ttl)
		}

		userTTL, err := client.TTL(
			context.Background(),
			userSessionsKey(sess.UserID()),
		).Result()
		if err != nil {
			t.Fatalf("failed to get user session TTL: %v", err)
		}

		if userTTL <= 0 || userTTL > duration {
			t.Fatalf(
				"expected user session TTL between 0 and %v, got %v",
				duration,
				userTTL,
			)
		}
	})

	t.Run("returns ErrSessionNotFound for missing session", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, time.Hour)

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

	t.Run("returns ErrSessionNotFound when session expires", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, 50*time.Millisecond)
		sess := newTestSession(t, "session-1", "user-1")

		if err := store.Add(context.Background(), sess); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		_, err := store.Get(context.Background(), sess.ID())
		if !errors.Is(err, appsession.ErrSessionNotFound) {
			t.Fatalf(
				"expected ErrSessionNotFound after expiry, got %v",
				err,
			)
		}
	})

	t.Run("lists sessions by user", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, time.Hour)

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

		foundFirst := false
		foundSecond := false

		for _, sess := range sessions {
			switch sess.ID() {
			case first.ID():
				foundFirst = true
			case second.ID():
				foundSecond = true
			}
		}

		if !foundFirst {
			t.Fatal("expected first session in result")
		}

		if !foundSecond {
			t.Fatal("expected second session in result")
		}
	})

	t.Run("returns nil when user has no sessions", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, time.Hour)

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

	t.Run("lists only sessions belonging to requested user", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, time.Hour)

		userOne := newTestSession(t, "session-1", "user-1")
		userTwo := newTestSession(t, "session-2", "user-2")

		if err := store.Add(context.Background(), userOne); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		if err := store.Add(context.Background(), userTwo); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		sessions, err := store.ListByUser(
			context.Background(),
			userOne.UserID(),
		)
		if err != nil {
			t.Fatalf("ListByUser() returned error: %v", err)
		}

		if len(sessions) != 1 {
			t.Fatalf("expected 1 session, got %d", len(sessions))
		}

		if sessions[0].ID() != userOne.ID() {
			t.Fatal("returned session belongs to the wrong user")
		}
	})

	t.Run("removes expired sessions from user session set", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, 50*time.Millisecond)
		sess := newTestSession(t, "session-1", "user-1")

		if err := store.Add(context.Background(), sess); err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

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

		ids, err := client.SMembers(
			context.Background(),
			userSessionsKey(sess.UserID()),
		).Result()
		if err != nil {
			t.Fatalf("failed to read user session set: %v", err)
		}

		if len(ids) != 0 {
			t.Fatalf(
				"expected expired session ID to be removed, got %v",
				ids,
			)
		}
	})

	t.Run("deletes existing session", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, time.Hour)
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

		ids, err := client.SMembers(
			context.Background(),
			userSessionsKey(sess.UserID()),
		).Result()
		if err != nil {
			t.Fatalf("failed to read user session set: %v", err)
		}

		if len(ids) != 0 {
			t.Fatalf("expected empty user session set, got %v", ids)
		}
	})

	t.Run("deletes one session while keeping other sessions", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, time.Hour)

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
			t.Fatalf("expected 1 remaining session, got %d", len(sessions))
		}

		if sessions[0].ID() != second.ID() {
			t.Fatalf("expected remaining session %q, got %q",
				second.ID(),
				sessions[0].ID(),
			)
		}
	})

	t.Run("delete is idempotent", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, time.Hour)

		id, err := appsession.NewSessionID("missing-session")
		if err != nil {
			t.Fatalf("failed to create session ID: %v", err)
		}

		if err := store.Delete(context.Background(), id); err != nil {
			t.Fatalf("Delete() returned error: %v", err)
		}

		if err := store.Delete(context.Background(), id); err != nil {
			t.Fatalf("second Delete() returned error: %v", err)
		}
	})

	t.Run("returns error when stored session contains invalid JSON", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, time.Hour)

		sess := newTestSession(t, "session-1", "user-1")

		err := client.Set(
			context.Background(),
			sessionKey(sess.ID()),
			"{invalid-json",
			time.Hour,
		).Err()
		if err != nil {
			t.Fatalf("failed to seed invalid session: %v", err)
		}

		_, err = store.Get(context.Background(), sess.ID())
		if err == nil {
			t.Fatal("expected Get() to return an error")
		}

		if !contains(err.Error(), "session.RedisStore.Get: unmarshal") {
			t.Fatalf("expected unmarshal error, got %v", err)
		}
	})

	t.Run("returns error when listed session contains invalid JSON", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, time.Hour)

		sess := newTestSession(t, "session-1", "user-1")

		err := client.SAdd(
			context.Background(),
			userSessionsKey(sess.UserID()),
			sess.ID().String(),
		).Err()
		if err != nil {
			t.Fatalf("failed to seed user session set: %v", err)
		}

		err = client.Expire(
			context.Background(),
			userSessionsKey(sess.UserID()),
			time.Hour,
		).Err()
		if err != nil {
			t.Fatalf("failed to set user session set expiry: %v", err)
		}

		err = client.Set(
			context.Background(),
			sessionKey(sess.ID()),
			"{invalid-json",
			time.Hour,
		).Err()
		if err != nil {
			t.Fatalf("failed to seed invalid session: %v", err)
		}

		_, err = store.ListByUser(
			context.Background(),
			sess.UserID(),
		)
		if err == nil {
			t.Fatal("expected ListByUser() to return an error")
		}

		if !contains(
			err.Error(),
			"session.RedisStore.ListByUser: unmarshal",
		) {
			t.Fatalf("expected unmarshal error, got %v", err)
		}
	})

	t.Run("returns error when deleting corrupted session", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, time.Hour)

		sess := newTestSession(t, "session-1", "user-1")

		err := client.Set(
			context.Background(),
			sessionKey(sess.ID()),
			"{invalid-json",
			time.Hour,
		).Err()
		if err != nil {
			t.Fatalf("failed to seed invalid session: %v", err)
		}

		err = store.Delete(context.Background(), sess.ID())
		if err == nil {
			t.Fatal("expected Delete() to return an error")
		}

		if !contains(
			err.Error(),
			"session.RedisStore.Delete: unmarshal",
		) {
			t.Fatalf("expected unmarshal error, got %v", err)
		}
	})

	t.Run("returns context error when context is cancelled", func(t *testing.T) {
		resetRedis(t, client)

		store := NewRedisStore(client, time.Hour)
		sess := newTestSession(t, "session-1", "user-1")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := store.Add(ctx, sess)
		if err == nil {
			t.Fatal("expected Add() to return an error")
		}

		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	})
}

func contains(value string, substring string) bool {
	return len(value) >= len(substring) &&
		(value == substring ||
			len(value) > len(substring) &&
				containsSubstring(value, substring))
}

func containsSubstring(value string, substring string) bool {
	for i := 0; i+len(substring) <= len(value); i++ {
		if value[i:i+len(substring)] == substring {
			return true
		}
	}

	return false
}
