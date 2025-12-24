package service

import (
	"context"
	"encoding/json"
	"errors"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"mizu/internal/dto/common"
	dto "mizu/internal/dto/contract"
	"mizu/internal/email"
	emlPrms "mizu/internal/email/params"
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
	contSt  store.ContractStore
	emlSndr email.EmailSender
}

// NewContractService constructs a ContractService that
// handles contract-related business logic. It needs a non-nil logger
// for audit and debugging, and a ContractStore for persistence.
func NewContractService(
	lg logger.Logger,
	evtSndr *event.EventSender,
	contSt store.ContractStore,
	emlSndr email.EmailSender) *ContractService {

	return &ContractService{
		lg:      lg,
		evtSndr: evtSndr,
		contSt:  contSt,
		emlSndr: emlSndr,
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
func (s *ContractService) CreateContract(
	ctx context.Context,
	projectId int64,
	request *dto.ContractCreateRequest) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"project_id", projectId,
			"err", err)
		return ErrInternalError
	}

	result, err := s.contSt.CreateOne(ctx, projectId, &params.ContractCreate{
		Name:     request.Name,
		Version:  request.Version,
		Contract: request.Contract,
	})

	if err != nil {
		if errors.Is(err, store.ErrForeignKeyViolation) {
			s.lg.Error("contract foreign key violated",
				"event", event.EventCreateFailed,
				"correlation_id", correlationId,
				"scope", "contract_service",
				"project_id", projectId,
				"actor_id", actor.Id,
				"err", err)
			return ErrBadRequest
		}

		if errors.Is(err, store.ErrUniqueViolation) {
			s.lg.Error("contract already exists in the project",
				"event", event.EventAlreadyExists,
				"correlation_id", correlationId,
				"scope", "contract_service",
				"project_id", projectId,
				"actor_id", actor.Id,
				"err", err)
			return ErrAlreadyExists
		}

		if errors.Is(err, store.ErrNotNullViolation) {
			s.lg.Error("contract not-null constraint violated",
				"event", event.EventCreateFailed,
				"correlation_id", correlationId,
				"scope", "contract_service",
				"project_id", projectId,
				"actor_id", actor.Id,
				"err", err)
			return ErrBadRequest
		}

		s.lg.Error("failed to create contract",
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
			s.lg.Error("failed to marshal message",
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

		s.evtSndr.SendTo(string(event.EventMessage), msgBytes, userIDs)

	}

	s.lg.Info("contract created successfully",
		"event", event.EventCreateSuccess,
		"correlation_id", correlationId,
		"scope", "contract_service",
		"project_id", projectId,
		"contract_id", result.ContractId,
		"actor_id", actor.Id)

	return nil
}

// SignContract method attempts to sign a contract version.
// If the contract version hasn't been signed by the requesting user,
// it sends a contract-signed email to relevant users.
//
//   - If the contract version is already signed, it returns service.ErrAlreadyExists
//   - If an error occurs, it returns service.ErrInternalError
func (s *ContractService) SignContractVersion(ctx context.Context, versionId int64) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"version_id", versionId,
			"err", err)
		return ErrInternalError
	}

	isAlreadySigned, err := s.contSt.SignVersion(ctx, versionId, actor.Id)
	if err != nil {
		s.lg.Error("failed to sign contract version",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"version_id", versionId,
			"err", err)
		return ErrInternalError
	}

	if isAlreadySigned {
		return ErrAlreadyExists
	}

	// Send contract-signed emails for new signs.
	signs, err := s.contSt.GetUsersWithSignature(ctx, versionId)
	if err != nil {
		s.lg.Error("failed get users with signatures for contract version",
			"event", event.EventGetFailed,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"version_id", versionId,
			"err", err)
	}

	if signs != nil {

		emailParams := emlPrms.ContractSignedRequest{
			ContractName:    signs.ContractName,
			ContractVersion: signs.ContractVersion,
			ContractText:    signs.ContractText,
			Signatures:      make([]emlPrms.UserSignature, 0),
		}

		for _, sign := range signs.Signatures {
			emailParams.Signatures = append(emailParams.Signatures, emlPrms.UserSignature{
				FirstName: sign.FirstName,
				LastName:  sign.LastName,
				Email:     sign.Email,
				SignedAt:  sign.SignedAt,
				Status:    sign.Status,
			})
		}

		if err := s.emlSndr.SendContractSigned(ctx, &emailParams); err != nil {
			s.lg.Warn("failed to send contract signed email",
				"event", event.EventEmailSendFailed,
				"correlation_id", correlationId,
				"scope", "user_service",
				"actor_id", actor.Id,
				"version_id", versionId,
				"err", err)
		}
	}

	return nil
}

// RejectContractVersion method attempts to reject a contract version.
// If the contract version hasn't been rejected by the requesting user,
// it sends a contract-rejected email to relevant users.
//
//   - If the contract version is already signed/rejected, it returns service.ErrAlreadyExists
//   - If an error occurs, it returns service.ErrInternalError
func (s *ContractService) RejectContractVersion(ctx context.Context, versionId int64) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"version_id", versionId,
			"err", err)
		return ErrInternalError
	}

	isAlreadySigned, err := s.contSt.RejectVersion(ctx, versionId, actor.Id)
	if err != nil {
		s.lg.Error("failed to reject contract version",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"version_id", versionId,
			"err", err)
		return ErrInternalError
	}

	if isAlreadySigned {
		return ErrAlreadyExists
	}

	// Send contract-signed emails for new rejections.
	signs, err := s.contSt.GetUsersWithSignature(ctx, versionId)
	if err != nil {
		s.lg.Error("failed get users with signatures for contract version",
			"event", event.EventGetFailed,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"version_id", versionId,
			"err", err)
	}

	if signs != nil {

		emailParams := emlPrms.ContractRejectedRequest{
			ContractName:    signs.ContractName,
			ContractVersion: signs.ContractVersion,
			ContractText:    signs.ContractText,
			Signatures:      make([]emlPrms.UserSignature, 0),
		}

		for _, sign := range signs.Signatures {
			emailParams.Signatures = append(emailParams.Signatures, emlPrms.UserSignature{
				FirstName: sign.FirstName,
				LastName:  sign.LastName,
				Email:     sign.Email,
				SignedAt:  sign.SignedAt,
				Status:    sign.Status,
			})
		}

		if err := s.emlSndr.SendContractRejected(ctx, &emailParams); err != nil {
			s.lg.Warn("failed to send contract rejected email",
				"event", event.EventEmailSendFailed,
				"correlation_id", correlationId,
				"scope", "user_service",
				"actor_id", actor.Id,
				"version_id", versionId,
				"err", err)
		}
	}

	return nil
}

// CreateRevision creates a contract revision for the specified contract.
// This method expects middleware to properly authorize requests.
//
//   - If request is considered invalid, it returns service.ErrBadRequest
//   - If an error occurs, it returns service.ErrInternalError
func (s *ContractService) CreateRevision(
	ctx context.Context,
	contractId int64,
	request *dto.RevisionCreateRequest) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"contract_id", contractId,
			"err", err)
		return ErrInternalError
	}

	err = s.contSt.CreateRevision(ctx, &params.ContractRevisionCreate{
		ContractId:  contractId,
		UserId:      actor.Id,
		Title:       request.Title,
		Description: request.Description,
	})

	if err != nil {
		if errors.Is(err, store.ErrForeignKeyViolation) {
			s.lg.Error("contract revision foreign key violated",
				"event", event.EventCreateFailed,
				"correlation_id", correlationId,
				"scope", "contract_service",
				"contract_id", contractId,
				"actor_id", actor.Id,
				"err", err)
			return ErrBadRequest
		}

		if errors.Is(err, store.ErrNotNullViolation) {
			if errors.Is(err, store.ErrNotNullViolation) {
				s.lg.Error("contract revision not-null constraint violated",
					"event", event.EventCreateFailed,
					"correlation_id", correlationId,
					"scope", "contract_service",
					"contract_id", contractId,
					"actor_id", actor.Id,
					"err", err)
				return ErrBadRequest
			}
		}

		s.lg.Error("failed to create contract revision",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"contract_id", contractId,
			"actor_id", actor.Id,
			"err", err)
		return ErrInternalError

	}

	return nil
}

// AcceptRevision accepts a contract revision specified by the revisionId parameter.
// Requesting user is assigned as the revision accepting user.
// This method expects middleware to properly authorize requests.
//
//   - If an error occurs, it returns service.ErrInternalError
func (s *ContractService) AcceptRevision(ctx context.Context, revisionId int64) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"revision_id", revisionId,
			"err", err)
		return ErrInternalError
	}

	if err := s.contSt.AcceptRevision(ctx, revisionId, actor.Id); err != nil {
		s.lg.Error("failed to accept contract revision",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"revision_id", revisionId,
			"actor_id", actor.Id,
			"err", err)
		return ErrInternalError
	}

	return nil
}

// RejectRevision rejects a contract revision specified by the revisionId parameter.
// Requesting user is assigned as the revision rejecting user.
// This method expects middleware to properly authorize requests.
//
//   - If an error occurs, it returns service.ErrInternalError
func (s *ContractService) RejectRevision(ctx context.Context, revisionId int64) error {
	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"revision_id", revisionId,
			"err", err)
		return ErrInternalError
	}

	if err := s.contSt.RejectRevision(ctx, revisionId, actor.Id); err != nil {
		s.lg.Error("failed to reject contract revision",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"revision_id", revisionId,
			"actor_id", actor.Id,
			"err", err)
		return ErrInternalError
	}

	return nil
}

func (s *ContractService) GetRevisions(
	ctx context.Context,
	query *dto.RevisionSearchQuery) (*common.Page[[]dto.RevisionResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"status", query.Status,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	revisions, err := s.contSt.GetRevisionsByUser(ctx, actor.Id, &params.ContractRevisionSearch{
		Keyword: query.Keyword,
		Status:  query.Status,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
	})

	if err != nil {
		s.lg.Error("could not get revision list",
			"event", event.EventGetFailed,
			"scope", "contract_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"status", query.Status,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	count, err := s.contSt.CountRevisionsByUser(ctx, actor.Id, &params.ContractRevisionSearch{
		Keyword: query.Keyword,
		Status:  query.Status,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
	})

	if err != nil {
		s.lg.Error("could not get revision count",
			"event", event.EventGetFailed,
			"scope", "contract_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"status", query.Status,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	revResp := make([]dto.RevisionResponse, len(revisions))

	for i, revision := range revisions {
		revResp[i] = dto.RevisionResponse{
			Id:           revision.Id,
			ContractId:   revision.ContractId,
			ContractName: revision.ContractName,
			Title:        revision.Title,
			Description:  revision.Description,
			CreatedAt:    revision.CreatedAt,
			UpdatedAt:    revision.UpdatedAt,
			Status:       revision.Status,
			ReqUser: dto.RevisionUser{
				FirstName: revision.ReqUserFirstName,
				LastName:  revision.ReqUserLastName,
			},
		}

		if revision.ResUserFirstName != nil && revision.ResUserLastName != nil {
			revResp[i].ResUser = &dto.RevisionUser{
				FirstName: *revision.ResUserFirstName,
				LastName:  *revision.ResUserLastName,
			}
		}
	}

	resp := common.Page[[]dto.RevisionResponse]{
		Count: count,
		Limit: query.Limit,
		Page:  query.Page,
		Data:  revResp,
	}

	return &resp, nil
}

// GetRevisionsByContract retrieves a paginated list of revisions for the specified contract.
// The contract is specified by the ID.
// It supports keyword and status filtering,
// and returns results wrapped in a common.Page payload.
// This method expects middleware to properly authorize requests.
//
//   - If an error occurs, it returns service.ErrInternalError
func (s *ContractService) GetRevisionsByContract(
	ctx context.Context,
	contractId int64,
	query *dto.RevisionSearchQuery) (*common.Page[[]dto.RevisionResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"contract_id", contractId,
			"status", query.Status,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	revisions, err := s.contSt.GetRevisionsByContractId(ctx, contractId, &params.ContractRevisionSearch{
		Keyword: query.Keyword,
		Status:  query.Status,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
	})

	if err != nil {
		s.lg.Error("could not get revision list",
			"event", event.EventGetFailed,
			"scope", "contract_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"status", query.Status,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	revResp := make([]dto.RevisionResponse, len(revisions.Items))

	for i, revision := range revisions.Items {
		revResp[i] = dto.RevisionResponse{
			Id:           revision.Id,
			ContractId:   revision.ContractId,
			ContractName: revision.ContractName,
			Title:        revision.Title,
			Description:  revision.Description,
			CreatedAt:    revision.CreatedAt,
			UpdatedAt:    revision.UpdatedAt,
			Status:       revision.Status,
			ReqUser: dto.RevisionUser{
				FirstName: revision.ReqUserFirstName,
				LastName:  revision.ReqUserLastName,
			},
		}

		if revision.ResUserFirstName != nil && revision.ResUserLastName != nil {
			revResp[i].ResUser = &dto.RevisionUser{
				FirstName: *revision.ResUserFirstName,
				LastName:  *revision.ResUserLastName,
			}
		}
	}

	resp := common.Page[[]dto.RevisionResponse]{
		Count: revisions.Total,
		Limit: query.Limit,
		Page:  query.Page,
		Data:  revResp,
	}

	return &resp, nil
}

func (s *ContractService) GetRevisionById(
	ctx context.Context,
	revisionId int64) (*dto.RevisionResponse, error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"revision_id", revisionId,
			"err", err)
		return nil, ErrInternalError
	}

	revision, err := s.contSt.GetRevisionById(ctx, revisionId)

	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			s.lg.Error("revision not found",
				"event", event.EventGetFailed,
				"scope", "contract_service",
				"correlation_id", correlationId,
				"actor_id", actor.Id,
				"revision_id", revisionId,
				"err", err)

			return nil, ErrNotFound
		}
		s.lg.Error("could not get revision list",
			"event", event.EventGetFailed,
			"scope", "contract_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"revision_id", revisionId,
			"err", err)
		return nil, ErrInternalError
	}

	resp := dto.RevisionResponse{
		Id:           revision.Id,
		ContractId:   revision.ContractId,
		ContractName: revision.ContractName,
		Project: dto.RevisionProject{
			Id:   revision.ProjectId,
			Name: revision.ProjectName,
		},
		Title:       revision.Title,
		Description: revision.Description,
		CreatedAt:   revision.CreatedAt,
		UpdatedAt:   revision.UpdatedAt,
		Status:      revision.Status,
		ReqUser: dto.RevisionUser{
			FirstName: revision.ReqUserFirstName,
			LastName:  revision.ReqUserLastName,
		},
	}

	if revision.ResUserFirstName != nil && revision.ResUserLastName != nil {
		resp.ResUser = &dto.RevisionUser{
			FirstName: *revision.ResUserFirstName,
			LastName:  *revision.ResUserLastName,
		}
	}

	return &resp, nil
}

// GetContractsByProject retrieves a paginated list of contracts for
// the specified project. The project is specified by the ID.
// It supports keyword and status filtering,
// and returns results wrapped in a common.Page payload.
// This method expects middleware to properly authorize requests.
//
//   - If an error occurs, it returns service.ErrInternalError
func (s *ContractService) GetContractsByProject(
	ctx context.Context,
	projectId int64,
	query *dto.ContractSearchQuery) (*common.Page[[]dto.ContractStatsResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"project_id", projectId,
			"status", query.Status,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	contracts, err := s.contSt.GetContractStatsByProject(ctx, projectId, actor.Id, &params.ContractSearch{
		Keyword: query.Keyword,
		Status:  query.Status,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
	})

	if err != nil {
		s.lg.Error("could not get contracts list",
			"event", event.EventGetFailed,
			"scope", "contract_service",
			"correlation_id", correlationId,
			"project_id", projectId,
			"actor_id", actor.Id,
			"status", query.Status,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	contractResp := make([]dto.ContractStatsResponse, len(contracts.Items))

	for i, contract := range contracts.Items {
		contractResp[i] = dto.ContractStatsResponse{
			Id:                contract.Id,
			Name:              contract.Name,
			Status:            contract.Status,
			CreatedAt:         contract.CreatedAt,
			Versions:          contract.Versions,
			Revisions:         contract.Revisions,
			AcceptedRevisions: contract.AcceptedRevisions,
			LatestVersion: dto.LatestVersion{
				Id:      contract.LatestVersionId,
				Version: contract.LatestVersion,
			},
			UserSignature: contract.UserSignature,
		}
	}

	resp := common.Page[[]dto.ContractStatsResponse]{
		Count: contracts.Total,
		Limit: query.Limit,
		Page:  query.Page,
		Data:  contractResp,
	}

	return &resp, nil
}

func (s *ContractService) GetVersions(
	ctx context.Context,
	contractId int64) ([]dto.VersionResponse, error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"contract_id", contractId,
			"err", err)
		return nil, ErrInternalError
	}

	versions, err := s.contSt.GetVersionsByContractId(ctx, contractId)
	if err != nil {
		s.lg.Error("could not get versions by contract id",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"contract_id", contractId,
			"actor_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp := make([]dto.VersionResponse, len(versions))

	for i, version := range versions {
		resp[i] = dto.VersionResponse{
			Id:        version.Id,
			CreatedAt: version.CreatedAt,
			Status:    version.Status,
			Version:   version.Version,
			Contract:  version.Contract,
		}
	}

	return resp, nil
}

func (s *ContractService) GetVersionSignatures(
	ctx context.Context,
	versionId int64) ([]dto.SignatureResponse, error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"version_id", versionId,
			"err", err)
		return nil, ErrInternalError
	}

	users, err := s.contSt.GetUsersWithSignature(ctx, versionId)
	if err != nil {
		s.lg.Error("could not get signatures by version id",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"version_id", versionId,
			"actor_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp := make([]dto.SignatureResponse, len(users.Signatures))

	for i, signature := range users.Signatures {
		resp[i] = dto.SignatureResponse{
			Id:        signature.Id,
			FirstName: signature.FirstName,
			LastName:  signature.LastName,
			SignedAt:  signature.SignedAt,
			Status:    signature.Status,
			Image:     signature.Image,
		}
	}

	return resp, nil
}

func (s *ContractService) GetByCurrentUser(
	ctx context.Context,
	query *dto.ContractSearchQuery) (*common.Page[[]dto.ContractStatsResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"err", err)
		return nil, ErrInternalError
	}

	search := params.ContractSearch{
		Keyword: query.Keyword,
		Status:  query.Status,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
	}

	contracts, err := s.contSt.GetByUserId(ctx, actor.Id, &search)
	if err != nil {
		s.lg.Error("could not get contracts of signed in user",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"actor_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	count, err := s.contSt.CountByUserId(ctx, actor.Id, &search)
	if err != nil {
		s.lg.Error("could not count contracts of signed in user",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"actor_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp := common.Page[[]dto.ContractStatsResponse]{
		Count: count,
		Page:  query.Page,
		Limit: query.Limit,
		Data:  make([]dto.ContractStatsResponse, len(contracts)),
	}

	for i, v := range contracts {
		resp.Data[i] = dto.ContractStatsResponse{
			Id:                v.Id,
			Name:              v.Name,
			ProjectName:       v.ProjectName,
			Status:            v.Status,
			CreatedAt:         v.CreatedAt,
			Versions:          v.Versions,
			Revisions:         v.Revisions,
			AcceptedRevisions: v.AcceptedRevisions,
			LatestVersion: dto.LatestVersion{
				Id:      v.LatestVersionId,
				Version: v.LatestVersion,
			},
			UserSignature: v.UserSignature,
		}
	}

	return &resp, nil
}

func (s *ContractService) GetStatById(
	ctx context.Context,
	contractId int64) (*dto.ContractStatsResponse, error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"err", err)
		return nil, ErrInternalError
	}

	stat, err := s.contSt.GetStatById(ctx, contractId)
	if err != nil {
		s.lg.Error("could not stats by contract id",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "contract_service",
			"actor_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp := dto.ContractStatsResponse{
		Id:                stat.Id,
		Name:              stat.Name,
		ProjectName:       stat.ProjectName,
		Status:            stat.Status,
		CreatedAt:         stat.CreatedAt,
		Versions:          stat.Versions,
		Revisions:         stat.Revisions,
		AcceptedRevisions: stat.AcceptedRevisions,
	}

	return &resp, nil
}
