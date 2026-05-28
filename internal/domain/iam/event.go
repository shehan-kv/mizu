package iam

import (
	"mizu/internal/domain/common"
	"time"
)

var (
	EventTypeUserCreated  common.EventType = "iam.user.created"
	EventTypeUserVerified common.EventType = "iam.user.verified"
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
func (e UserCreatedEvent) EventScope() common.EventScope {
	return common.EventScopeInternal
}

type UserVerifiedEvent struct {
	UserID     UserID
	OccurredAt time.Time
}

func (e UserVerifiedEvent) EventType() common.EventType {
	return EventTypeUserVerified
}
func (e UserVerifiedEvent) EventScope() common.EventScope {
	return common.EventScopeInternal
}
