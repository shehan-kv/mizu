package verification

import (
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"time"
)

var (
	EventTypeVerificationCreated common.EventType = "verification.created"
)

type VerificationCreatedEvent struct {
	UserID         iam.UserID
	VerificationID VerificationID
	OccurredAt     time.Time
}

func (e VerificationCreatedEvent) EventType() common.EventType {
	return EventTypeVerificationCreated
}
func (e VerificationCreatedEvent) EventScope() common.EventScope {
	return common.EventScopeInternal
}
