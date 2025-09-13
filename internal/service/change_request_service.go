package service

import (
	"context"
	"errors"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	dto "mizu/internal/dto/change_request"
	"mizu/internal/event"
	"mizu/internal/logger"
	"mizu/internal/middleware"
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

func (chngReqSrv *ChangeRequestService) Create(
	ctx context.Context,
	projectId int64,
	request *dto.ChangeReqCreateRequest) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		chngReqSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"project_id", projectId,
			"err", err)
		return ErrInternalError
	}

	_, err = chngReqSrv.chngReqSt.CreateOne(ctx, actor.Id, &params.ChangeRequestCreate{
		ProjectId: projectId,
		Title:     request.Title,
		Content:   request.Content,
	})

	if err != nil {
		if errors.Is(err, store.ErrForeignKeyViolation) {
			return ErrBadRequest
		}

		if errors.Is(err, store.ErrNotNullViolation) {
			return ErrBadRequest
		}

		return ErrInternalError
	}

	return nil
}
