package session

import (
	"context"
	"encoding/json"
	"fmt"
	"mizu/internal/application/session"
	"mizu/internal/domain/iam"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore is a Redis-backed implementation of Store.
// Sessions are stored as JSON with an expiry matching the session duration.
// User-to-session mappings are stored as Redis sets for efficient lookup.
type RedisStore struct {
	client   *redis.Client
	duration time.Duration
}

func NewRedisStore(client *redis.Client, duration time.Duration) *RedisStore {
	return &RedisStore{
		client:   client,
		duration: duration,
	}
}

// sessionKey returns the Redis key for a session.
func sessionKey(id session.SessionID) string {
	return "session:" + id.String()
}

// userSessionsKey returns the Redis key for the set of session IDs belonging to a user.
func userSessionsKey(userID iam.UserID) string {
	return "user_sessions:" + userID.String()
}

// Add stores the session in Redis with an expiry.
// The user-to-session mapping is stored in a Redis set with the same expiry.
func (s *RedisStore) Add(ctx context.Context, sess *session.Session) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("session.RedisStore.Add: marshal: %w", err)
	}

	pipe := s.client.Pipeline()

	pipe.Set(ctx, sessionKey(sess.ID()), data, s.duration)
	pipe.SAdd(ctx, userSessionsKey(sess.UserID()), sess.ID().String())
	pipe.Expire(ctx, userSessionsKey(sess.UserID()), s.duration)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("session.RedisStore.Add: %w", err)
	}

	return nil
}

// Get retrieves a session by its ID.
// Returns ErrSessionNotFound if the session does not exist or has expired.
func (s *RedisStore) Get(ctx context.Context, id session.SessionID) (*session.Session, error) {
	data, err := s.client.Get(ctx, sessionKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, session.ErrSessionNotFound
		}
		return nil, fmt.Errorf("session.RedisStore.Get: %w", err)
	}

	var sess session.Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("session.RedisStore.Get: unmarshal: %w", err)
	}

	return &sess, nil
}

// ListByUser returns all active sessions for the given user.
// Returns nil if the user has no active sessions.
func (s *RedisStore) ListByUser(ctx context.Context, userID iam.UserID) ([]*session.Session, error) {
	ids, err := s.client.SMembers(ctx, userSessionsKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("session.RedisStore.ListByUser: %w", err)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	// batch fetch all sessions in one round trip
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = "session:" + id
	}

	results, err := s.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("session.RedisStore.ListByUser: %w", err)
	}

	sessions := make([]*session.Session, 0, len(results))
	expiredIDs := make([]any, 0)

	for i, result := range results {
		if result == nil {
			// session expired, mark for removal from the set
			expiredIDs = append(expiredIDs, ids[i])
			continue
		}

		var sess session.Session
		if err := json.Unmarshal([]byte(result.(string)), &sess); err != nil {
			return nil, fmt.Errorf("session.RedisStore.ListByUser: unmarshal: %w", err)
		}

		sessions = append(sessions, &sess)
	}

	// clean up expired session IDs from the user set
	if len(expiredIDs) > 0 {
		s.client.SRem(ctx, userSessionsKey(userID), expiredIDs...)
	}

	if len(sessions) == 0 {
		return nil, nil
	}

	return sessions, nil
}

// Delete removes a session by its ID.
// Delete is idempotent, deleting a non-existent session is not an error.
func (s *RedisStore) Delete(ctx context.Context, id session.SessionID) error {
	data, err := s.client.Get(ctx, sessionKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return fmt.Errorf("session.RedisStore.Delete: get session: %w", err)
	}

	var sess session.Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return fmt.Errorf("session.RedisStore.Delete: unmarshal: %w", err)
	}

	pipe := s.client.Pipeline()
	pipe.Del(ctx, sessionKey(id))
	pipe.SRem(ctx, userSessionsKey(sess.UserID()), id.String())

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("session.RedisStore.Delete: %w", err)
	}

	return nil
}
