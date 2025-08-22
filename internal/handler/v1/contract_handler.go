package v1

import (
	"mizu/internal/db/store"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"mizu/internal/service"
	"mizu/internal/session"
	"net/http"
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
