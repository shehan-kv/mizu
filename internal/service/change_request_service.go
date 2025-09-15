package service

import (
	"context"
	"errors"
	"math"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	dto "mizu/internal/dto/change_request"
	"mizu/internal/dto/common"
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

func (chngReqSrv *ChangeRequestService) CreateEntry(
	ctx context.Context,
	requestId int64,
	request *dto.EntryCreateRequest) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		chngReqSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"change_request_id", requestId,
			"err", err)
		return ErrInternalError
	}

	err = chngReqSrv.chngReqSt.CreateEntry(ctx, actor.Id, requestId, request.Content)

	if err != nil {
		if errors.Is(err, store.ErrForeignKeyViolation) {
			return ErrBadRequest
		}

		if errors.Is(err, store.ErrUnexpectedType) {
			return ErrAlreadyExists
		}

		if errors.Is(err, store.ErrNotNullViolation) {
			return ErrBadRequest
		}

		return ErrInternalError
	}

	return nil
}

func (chngReqSrv *ChangeRequestService) GetAllByProject(
	ctx context.Context,
	projectId int64,
	query *dto.ChangeReqSearch) (*common.Page[[]dto.ChangeReqResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		chngReqSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "change_request_service",
			"project_id", actor,
			"err", err)
		return nil, ErrInternalError
	}

	result, err := chngReqSrv.chngReqSt.GetByProjectId(ctx, projectId, &params.ChangeRequestSearch{
		Keyword: query.Keyword,
		Status:  query.Status,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
	})

	if err != nil {
		return nil, ErrInternalError
	}

	numOfPages := math.Ceil(float64(result.Total) / float64(query.Limit))
	reqResp := common.Page[[]dto.ChangeReqResponse]{
		CurrentPage: query.Page,
		Limit:       query.Limit,
		TotalPages:  int64(numOfPages),
		Data:        make([]dto.ChangeReqResponse, len(result.Items)),
	}

	for i, item := range result.Items {
		reqResp.Data[i] = dto.ChangeReqResponse{
			Id:        item.Id,
			Title:     item.Title,
			CreatedAt: item.CreatedAt,
			RequestedBy: &dto.ChangeReqUserResponse{
				Id:        item.ReqUserId,
				FirstName: item.ReqUserFirstName,
				LastName:  item.ReqUserLastName,
			},
			Project: &dto.ChangeReqProjectResponse{
				Id:   item.ProjectId,
				Name: item.ProjectName,
			},
			Status: item.Status,
		}
	}

	return &reqResp, nil
}

func (chngReqSrv *ChangeRequestService) CloseById(ctx context.Context, requestId int64) error {
	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		chngReqSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "change_request_service",
			"project_id", actor,
			"err", err)
		return ErrInternalError
	}

	err = chngReqSrv.chngReqSt.CloseById(ctx, requestId)
	if err != nil {
		if errors.Is(err, store.ErrUnexpectedType) {
			return ErrAlreadyExists
		}

		return ErrInternalError
	}

	return nil
}

func (chngReqSrv *ChangeRequestService) GetById(ctx context.Context, requestId int64) (*dto.ChangeReqDetailsResponse, error) {
	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		chngReqSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "change_request_service",
			"project_id", actor,
			"err", err)
		return nil, ErrInternalError
	}

	req, err := chngReqSrv.chngReqSt.GetById(ctx, requestId)
	if err != nil {
		return nil, ErrInternalError
	}

	entries, err := chngReqSrv.chngReqSt.GetEntriesByRequestId(ctx, requestId)
	if err != nil {
		return nil, ErrInternalError
	}

	reqDetails := dto.ChangeReqDetailsResponse{
		Id:        req.Id,
		Title:     req.Title,
		CreatedAt: req.CreatedAt,
		RequestedBy: &dto.ChangeReqUserResponse{
			Id:        req.ReqUserId,
			FirstName: req.ReqUserFirstName,
			LastName:  req.ReqUserLastName,
		},
		Project: &dto.ChangeReqProjectResponse{
			Id:   req.ProjectId,
			Name: req.ProjectName,
		},
		Status:  req.Status,
		Entries: make([]dto.ChangeReqEntryResponse, len(entries)),
	}

	for i, entry := range entries {
		reqDetails.Entries[i] = dto.ChangeReqEntryResponse{
			Id:        entry.Id,
			CreatedAt: entry.CreatedAt,
			Content:   entry.Content,
			User: dto.ChangeReqUserResponse{
				Id:        entry.UserId,
				FirstName: entry.UserFirstName,
				LastName:  entry.UserLastName,
			},
		}
	}

	return &reqDetails, nil
}
