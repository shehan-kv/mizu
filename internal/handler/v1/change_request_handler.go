package v1

import (
	"encoding/json"
	"errors"
	"mizu/internal/db/store"
	dto "mizu/internal/dto/change_request"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"mizu/internal/service"
	"mizu/internal/session"
	"net/http"
	"strconv"
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

	mux.Handle("POST /{projectId}", mwChain.Handle(chngReqHndl.CreateRequest))

	return mux
}

func (chngReqHndl *ChangeRequestHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {

	prjId := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(prjId, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var createRequest dto.ChangeReqCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ok := createRequest.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = chngReqHndl.chngReqSrv.Create(r.Context(), parsedPrjId, &createRequest)

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
