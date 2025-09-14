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
	mux.Handle("POST /entry/{requestId}", mwChain.Handle(chngReqHndl.CreateEntry))

	// eg, /change-requests/project/{projectId}
	mux.Handle("GET /project/{projectId}", mwChain.Handle(chngReqHndl.GetAllByProject))

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

func (chngReqHndl *ChangeRequestHandler) CreateEntry(w http.ResponseWriter, r *http.Request) {

	reqId := r.PathValue("requestId")
	parsedReqId, err := strconv.ParseInt(reqId, 10, 64)
	if err != nil || parsedReqId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var createRequest dto.EntryCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ok := createRequest.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = chngReqHndl.chngReqSrv.CreateEntry(r.Context(), parsedReqId, &createRequest)

	if err != nil {
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

func (chngReqHndl *ChangeRequestHandler) GetAllByProject(w http.ResponseWriter, r *http.Request) {

	prjId := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(prjId, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	keyword := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	strPage := r.URL.Query().Get("page")
	strLimit := r.URL.Query().Get("limit")

	var page int64
	var limit int64

	if strPage == "" {
		page = 1
	} else {
		parsedPage, err := strconv.ParseInt(strPage, 10, 64)
		if err != nil || parsedPage <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		page = parsedPage
	}

	if strLimit == "" {
		limit = 15
	} else {
		parsedLimit, err := strconv.ParseInt(strLimit, 10, 64)
		if err != nil || parsedLimit <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	result, err := chngReqHndl.chngReqSrv.GetAllByProject(r.Context(), parsedPrjId, &dto.ChangeReqSearch{
		Keyword: keyword,
		Status:  status,
		Page:    page,
		Limit:   limit,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
