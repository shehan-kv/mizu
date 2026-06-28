package iam

import (
	"mizu/internal/domain/common"
	"time"
)

var (
	EventTypeUserCreated      common.EventType = "iam.user.created"
	EventTypeUserVerified     common.EventType = "iam.user.verified"
	EventTypeUserEmailChanged common.EventType = "iam.user.email.changed"

	EventTypeRecoveryCreated common.EventType = "iam.recovery.created"
)

type UserCreatedEvent struct {
	UserID     UserID
	Name       Name
	Email      Email
	CreatedBy  UserID
	OccurredAt time.Time
}

func (e UserCreatedEvent) EventType() common.EventType {
	return EventTypeUserCreated
}

type UserVerifiedEvent struct {
	UserID     UserID
	OccurredAt time.Time
}

func (e UserVerifiedEvent) EventType() common.EventType {
	return EventTypeUserVerified
}

type UserEmailChangedEvent struct {
	UserID     UserID
	NewEmail   Email
	OccurredAt time.Time
}

func (e UserEmailChangedEvent) EventType() common.EventType {
	return EventTypeUserEmailChanged
}

type RecoveryCreatedEvent struct {
	UserID     UserID
	Token      RecoveryToken
	OccurredAt time.Time
}

func (e RecoveryCreatedEvent) EventType() common.EventType {
	return EventTypeUserEmailChanged
}
