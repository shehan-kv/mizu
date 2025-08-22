package service

import (
	"context"
	"encoding/json"
	"errors"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	dto "mizu/internal/dto/contract"
	"mizu/internal/event"
	"mizu/internal/logger"
	"mizu/internal/middleware"
)

// ContractService handles contract-related business logic.
// It relies on the provided ContractStore for database operations
// and uses the Logger for audit and debugging.
type ContractService struct {
	lg      logger.Logger
	evtSndr *event.EventSender
	usrSt   store.UserStore
	contSt  store.ContractStore
}

// NewContractService constructs a ContractService that
// handles contract-related business logic. It needs a non-nil logger
// for audit and debugging, and a ContractStore for persistence.
func NewContractService(
	lg logger.Logger,
	evtSndr *event.EventSender,
	usrSt store.UserStore,
	contSt store.ContractStore) *ContractService {

	return &ContractService{
		lg:      lg,
		evtSndr: evtSndr,
		usrSt:   usrSt,
		contSt:  contSt,
	}
}

// CreateContract takes a context.Context, a projectId and a pointer
// to dto.ContractCreateRequest and attempts to create a contract
// for a specified project. It expects middleware to handle authorization
// (checking if the user has access to create a contract for the specified project).
// This method sends the system generated message to connected users.
//
//   - If a request is considered invalid, it returns service.ErrBadRequest.
//   - If the project already contains a contract by the same name, it returns service.ErrAlreadyExists.
//   - If an internal error occurs, it returns service.ErrInternalError.
func (contSrv *ContractService) CreateContract(
	ctx context.Context,
	projectId int64,
	request *dto.ContractCreateRequest) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		contSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"project_id", projectId,
			"err", err)
		return ErrInternalError
	}

	result, err := contSrv.contSt.CreateOne(ctx, projectId, &params.ContractCreate{
		Name:     request.Name,
		Version:  request.Version,
		Contract: request.Contract,
	})

	if err != nil {
		if errors.Is(err, store.ErrForeignKeyViolation) {
			contSrv.lg.Error("contract foreign key violated",
				"event", event.EventCreateFailed,
				"correlation_id", correlationId,
				"scope", "contract_service",
				"project_id", projectId,
				"actor_id", actor.Id,
				"err", err)
			return ErrBadRequest
		}

		if errors.Is(err, store.ErrUniqueViolation) {
			contSrv.lg.Error("contract already exists in the project",
				"event", event.EventAlreadyExists,
				"correlation_id", correlationId,
				"scope", "contract_service",
				"project_id", projectId,
				"actor_id", actor.Id,
				"err", err)
			return ErrAlreadyExists
		}

		if errors.Is(err, store.ErrNotNullViolation) {
			contSrv.lg.Error("contract not-null constraint violated",
				"event", event.EventCreateFailed,
				"correlation_id", correlationId,
				"scope", "contract_service",
				"project_id", projectId,
				"actor_id", actor.Id,
				"err", err)
			return ErrBadRequest
		}

		contSrv.lg.Error("failed to create contract",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"project_id", projectId,
			"actor_id", actor.Id,
			"err", err)
		return ErrInternalError
	}

	// try to send system generated messages to
	// users assigned to the relevant channels
	for channelId, msg := range result.Messages {

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			contSrv.lg.Error("failed to marshal message",
				"event", event.EventInternalError,
				"correlation_id", correlationId,
				"scope", "contract_service",
				"project_id", projectId,
				"message_id", msg.MessageId,
				"channel_id", channelId,
				"actor_id", actor.Id,
				"err", err)

			continue
		}

		// skip the current iteration if users or message is empty
		userIDs := result.UserIds[channelId]
		if len(userIDs) == 0 || len(msgBytes) == 0 {
			continue
		}

		contSrv.evtSndr.SendTo(string(event.EventMessage), msgBytes, userIDs)

	}

	contSrv.lg.Info("contract created successfully",
		"event", event.EventCreateSuccess,
		"correlation_id", correlationId,
		"scope", "contract_service",
		"project_id", projectId,
		"contract_id", result.ContractId,
		"actor_id", actor.Id)

	return nil
}
