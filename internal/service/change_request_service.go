package service

import (
	"mizu/internal/db/store"
	"mizu/internal/event"
	"mizu/internal/logger"
)

// ChangeRequestService handles change-request related business logic.
// It relies on the provided ChangeRequestStore for database operations
// and uses the Logger for audit and debugging.
type ChangeRequestService struct {
	lg        logger.Logger
	evtSndr   *event.EventSender
	chngReqSt store.ChangeRequestStore
}

// NewContractService constructs a ChangeRequestService that
// handles change-request related business logic. It needs a non-nil logger
// for audit and debugging, and a ContractStore for persistence.
func NewChangeRequestService(
	lg logger.Logger,
	evtSndr *event.EventSender,
	chngReqSt store.ChangeRequestStore) *ChangeRequestService {

	return &ChangeRequestService{
		lg:        lg,
		evtSndr:   evtSndr,
		chngReqSt: chngReqSt,
	}
}
