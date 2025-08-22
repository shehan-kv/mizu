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
