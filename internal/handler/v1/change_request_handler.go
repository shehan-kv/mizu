package v1

import (
	"mizu/internal/db/store"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"mizu/internal/service"
	"mizu/internal/session"
	"net/http"
)

// ChangeRequestHandler provides HTTP handlers
// for change-request related endpoints.
type ChangeRequestHandler struct {
	chngReqSrv *service.ChangeRequestService
}

// NewChangeRequestHandler constructs a new ChangeRequestHandler.
func NewChangeRequestHandler(chngReqSrv *service.ChangeRequestService) *ChangeRequestHandler {
	return &ChangeRequestHandler{
		chngReqSrv: chngReqSrv,
	}
}

// GetMux returns an http.ServeMux for the change-request routes and middleware.
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
func (chngReqHndl *ChangeRequestHandler) GetMux(
	lg logger.Logger,
	seSt session.SessionStore,
	usrSt store.UserStore) *http.ServeMux {

	mwChain := middleware.NewChain()
	mwChain.Add(
		middleware.CorrelationId(lg),
		middleware.Authenticated(lg, seSt, usrSt))

	mux := http.NewServeMux()

	// TODO: Define routes here

	return mux
}
