package contract

import (
	"mizu/internal/domain/iam"
	"time"
)

type Signatory struct {
	userID    iam.UserID
	status    SignatoryStatus
	updatedAt time.Time
}

func NewSignatory(userID iam.UserID, now time.Time) Signatory {
	return Signatory{
		userID:    userID,
		status:    SignatoryStatusPending,
		updatedAt: now,
	}
}

func RestoreSignatory(userID iam.UserID, status SignatoryStatus, updatedAt time.Time) Signatory {
	return Signatory{
		userID:    userID,
		status:    status,
		updatedAt: updatedAt,
	}
}

func (s Signatory) UserID() iam.UserID {
	return s.userID
}

func (s Signatory) Status() SignatoryStatus {
	return s.status
}

func (s Signatory) UpdatedAt() time.Time {
	return s.updatedAt
}
