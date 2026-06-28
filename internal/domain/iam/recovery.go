package iam

import (
	"mizu/internal/domain/common"
	"time"
)

type Recovery struct {
	userID UserID
	token  RecoveryToken

	events []common.Event

	version int

	createdAt time.Time
	expiresAt time.Time
}

func NewRecovery(
	userID UserID,
	token RecoveryToken,
	now time.Time,
) *Recovery {
	r := Recovery{
		userID:    userID,
		token:     token,
		version:   1,
		createdAt: now,
		expiresAt: now.Add(time.Hour),
	}

	r.events = append(r.events, RecoveryCreatedEvent{
		UserID:     userID,
		Token:      token,
		OccurredAt: now,
	})

	return &r
}

func RestoreRecovery(
	userID UserID,
	token RecoveryToken,
	version int,
	createdAt time.Time,
	expiresAt time.Time,
) *Recovery {
	return &Recovery{
		userID:    userID,
		token:     token,
		version:   version,
		createdAt: createdAt,
		expiresAt: expiresAt,
	}
}

func (r *Recovery) UserID() UserID {
	return r.userID
}

func (r *Recovery) Token() RecoveryToken {
	return r.token
}

func (r *Recovery) Version() int {
	return r.version
}

func (r *Recovery) CreatedAt() time.Time {
	return r.createdAt
}

func (r *Recovery) ExpiresAt() time.Time {
	return r.expiresAt
}

func (r *Recovery) IsExpired(now time.Time) bool {
	return now.After(r.expiresAt)
}

func (r *Recovery) PullEvents() []common.Event {
	events := r.events
	r.events = nil

	return events
}
