package v1

import (
	"encoding/json"
	"errors"
	"mizu/internal/db/store"
	dto "mizu/internal/dto/contract"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"mizu/internal/service"
	"mizu/internal/session"
	"net/http"
	"strconv"
)

// ContractHandler provides HTTP handlers for contract-related endpoints.
type ContractHandler struct {
	contSrv *service.ContractService
}

// NewContractHandler constructs a new ContractHandler.
func NewContractHandler(contSrv *service.ContractService) *ContractHandler {
	return &ContractHandler{
		contSrv: contSrv,
	}
}

// GetMux returns an http.ServeMux for the contract routes and middleware.
// It defines the routes and handler function for each route.
// Registers middleware for the routes.
//
// Parameters:
//   - lg: an implementation of logger.Logger
//   - seSt: an implementation of session.SessionStore
//   - usrSt: an implementation of store.UserStore
//
// Returns:
//   - a *http.ServeMux
func (contHndl *ContractHandler) GetMux(
	lg logger.Logger,
	seSt session.SessionStore,
	usrSt store.UserStore) *http.ServeMux {

	mwChain := middleware.NewChain()
	mwChain.Add(
		middleware.CorrelationId(lg),
		middleware.Authenticated(lg, seSt, usrSt))

	mux := http.NewServeMux()

	mux.Handle("POST /{projectId}", mwChain.Handle(contHndl.CreateContract))
	mux.Handle("POST /sign/{versionId}", mwChain.Handle(contHndl.SignContractVersion))
	mux.Handle("POST /reject/{versionId}", mwChain.Handle(contHndl.RejectContractVersion))
	mux.Handle("POST /revision/{contractId}", mwChain.Handle(contHndl.CreateContractRevision))
	mux.Handle("POST /revision/accept/{revisionId}", mwChain.Handle(contHndl.AcceptRevision))
	mux.Handle("POST /revision/reject/{revisionId}", mwChain.Handle(contHndl.RejectRevision))

	return mux
}

// CreateContract handles HTTP POST requests to create a new contract for
// a specified project. The project ID is expected as a path parameter, eg: contracts/{projectId}.
// If contract is successfully created, it returns HTTP 201 Created response.
// This function expects middleware to properly authorize requests.
// Expects a dto.ContractCreateRequest as request body.
//
//   - If projectId is missing or invalid, HTTP 400 BadRequest is returned
//   - If request body is invalid, HTTP 400 BadRequest is returned
//   - If contract already exists, HTTP 409 Conflict is returned
//   - If an internal error occurs, HTTP 500 InternalServerError is returned
func (contHndl *ContractHandler) CreateContract(w http.ResponseWriter, r *http.Request) {

	prjId := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(prjId, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var createRequest dto.ContractCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ok := createRequest.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := contHndl.contSrv.CreateContract(r.Context(), parsedPrjId, &createRequest); err != nil {
		if errors.Is(err, service.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// SignContractVersion handles signing of a contract version.
// The version ID is expected as a path parameter, eg: sign/{versionId}.
// If contract is successfully created, it returns HTTP 200 OK response.
// This function expects middleware to properly authorize requests.
//
//   - If versionId is missing or invalid, HTTP 400 BadRequest is returned
//   - If contract version already signed, HTTP 409 Conflict is returned
//   - If contract version is rejected, HTTP 409 Conflict is returned
//   - If an internal error occurs, HTTP 500 InternalServerError is returned
func (contHndl *ContractHandler) SignContractVersion(w http.ResponseWriter, r *http.Request) {

	versionId := r.PathValue("versionId")
	parsedVerId, err := strconv.ParseInt(versionId, 10, 64)
	if err != nil || parsedVerId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := contHndl.contSrv.SignContractVersion(r.Context(), parsedVerId); err != nil {
		if errors.Is(err, service.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RejectContractVersion handles rejecting a contract version.
// The version ID is expected as a path parameter, eg: sign/{versionId}.
// If contract is successfully rejected, it returns HTTP 200 OK response.
// This function expects middleware to properly authorize requests.
//
//   - If versionId is missing or invalid, HTTP 400 BadRequest is returned
//   - If contract version already signed, HTTP 409 Conflict is returned
//   - If contract version is already rejected, HTTP 409 Conflict is returned
//   - If an internal error occurs, HTTP 500 InternalServerError is returned
func (contHndl *ContractHandler) RejectContractVersion(w http.ResponseWriter, r *http.Request) {

	versionId := r.PathValue("versionId")
	parsedVerId, err := strconv.ParseInt(versionId, 10, 64)
	if err != nil || parsedVerId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := contHndl.contSrv.RejectContractVersion(r.Context(), parsedVerId); err != nil {
		if errors.Is(err, service.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CreateContractRevision handles creating a contract revision.
// The contract ID is expected as a path parameter, eg: {contractId}.
// If contract is successfully rejected, it returns HTTP 201 Created response.
// This method expects middleware to properly authorize requests.
//
//   - If contractId is missing or invalid, HTTP 400 BadRequest is returned
//   - If an internal error occurs, HTTP 500 InternalServerError is returned
func (contHndl *ContractHandler) CreateContractRevision(w http.ResponseWriter, r *http.Request) {
	contractId := r.PathValue("contractId")
	parsedContractId, err := strconv.ParseInt(contractId, 10, 64)
	if err != nil || parsedContractId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var createRequest dto.RevisionCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ok := createRequest.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = contHndl.contSrv.CreateRevision(
		r.Context(),
		parsedContractId,
		&createRequest)

	if err != nil {
		if errors.Is(err, service.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// AcceptRevision handles HTTP POST requests for accepting
// a contract revision by the user making the request.
// The revision ID is expected as a path parameter, eg: {revisionId}.
// If contract revision is successfully accepted, it returns HTTP 200 OK response.
// This method expects middleware to properly authorize requests.
//
//   - If revisionId is missing or invalid, HTTP 400 BadRequest is returned
//   - If an internal error occurs, HTTP 500 InternalServerError is returned
func (contHndl *ContractHandler) AcceptRevision(w http.ResponseWriter, r *http.Request) {
	revisionId := r.PathValue("revisionId")
	parsedRevisionId, err := strconv.ParseInt(revisionId, 10, 64)
	if err != nil || parsedRevisionId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := contHndl.contSrv.AcceptRevision(r.Context(), parsedRevisionId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RejectRevision handles HTTP POST requests for rejecting
// a contract revision by the user making the request.
// The revision ID is expected as a path parameter, eg: {revisionId}.
// This method expects middleware to properly authorize requests.
//
//   - If contract revision is successfully rejected, HTTP 200 is returned.
//   - If revisionId is missing or invalid, HTTP 400 BadRequest is returned.
//   - If an internal error occurs, HTTP 500 InternalServerError is returned.
func (contHndl *ContractHandler) RejectRevision(w http.ResponseWriter, r *http.Request) {
	revisionId := r.PathValue("revisionId")
	parsedRevisionId, err := strconv.ParseInt(revisionId, 10, 64)
	if err != nil || parsedRevisionId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := contHndl.contSrv.RejectRevision(r.Context(), parsedRevisionId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
