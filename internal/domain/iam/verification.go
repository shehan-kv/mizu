package iam

import (
	"mizu/internal/domain/common"
	"time"
)

type Verification struct {
	id     VerificationID
	userID UserID
	events []common.Event

	version   int
	createdAt time.Time
}

func NewVerification(id VerificationID, userID UserID, now time.Time) *Verification {

	v := Verification{
		id:        id,
		userID:    userID,
		version:   1,
		createdAt: now,
	}
	v.events = append(v.events, VerificationCreatedEvent{
		UserID:         userID,
		VerificationID: id,
		OccurredAt:     now,
	})

	return &v
}

func RestoreVerification(id VerificationID, userID UserID, version int, createdAt time.Time) *Verification {
	return &Verification{
		id:        id,
		userID:    userID,
		version:   version,
		createdAt: createdAt,
	}
}

func (v *Verification) ID() VerificationID {
	return v.id
}

func (v *Verification) UserID() UserID {
	return v.userID
}

func (v *Verification) Version() int {
	return v.version
}

func (v *Verification) CreatedAt() time.Time {
	return v.createdAt
}

func (v *Verification) PullEvents() []common.Event {
	events := v.events
	v.events = nil

	return events
}

func (v *Verification) Equals(verification *Verification) bool {
	return v.id == verification.ID()
}
