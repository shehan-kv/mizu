package v1

import (
	"encoding/json"
	"errors"
	"mizu/internal/db/store"
	dto "mizu/internal/dto/project"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"mizu/internal/service"
	"mizu/internal/session"
	"net/http"
	"strconv"
)

// Handles project-related HTTP requests.
//
// Uses an ProjectService to perform
// project operations
type ProjectHandler struct {
	prjSrv *service.ProjectService
}

// Creates a new instance of ProjectHandler
//
// Parameters:
//   - prjSrv: a pointer to a ProjectService
//
// Returns:
//   - a pointer to a new ProjectHandler
func NewProjectHandler(prjSrv *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		prjSrv: prjSrv,
	}
}

// Creates a ServeMux for the project routes and middleware.
// Defines the routes and handler function for each route.
// Registers middleware for the routes.
//
// Parameters:
//   - lg: an implementation of logger.Logger
//   - seSt: an implementation of session.SessionStore
//   - usrSt: an implementation of store.UserStore
//
// Returns:
//   - a *http.ServeMux
func (prjHndl *ProjectHandler) GetMux(
	lg logger.Logger,
	seSt session.SessionStore,
	usrSt store.UserStore) *http.ServeMux {

	mwChain := middleware.NewChain()
	mwChain.Add(
		middleware.CorrelationId(lg),
		middleware.Authenticated(lg, seSt, usrSt))

	mux := http.NewServeMux()

	mux.Handle("POST /", mwChain.Handle(prjHndl.CreateProject))
	mux.Handle("GET /", mwChain.Handle(prjHndl.GetProjects))
	mux.Handle("GET /{projectId}", mwChain.Handle(prjHndl.GetOneById))
	mux.Handle("POST /{projectId}/task", mwChain.Handle(prjHndl.CreateTask))
	mux.Handle("GET /{projectId}/task", mwChain.Handle(prjHndl.GetTasksByProject))
	mux.Handle("GET /{projectId}/task/metrics/complete", mwChain.Handle(prjHndl.GetTaskCompleteCountByProject))
	mux.Handle("POST /{projectId}/status/started", mwChain.Handle(prjHndl.SetStatusStarted))
	mux.Handle("POST /{projectId}/status/paused", mwChain.Handle(prjHndl.SetStatusPaused))
	mux.Handle("GET /metrics/create", mwChain.Handle(prjHndl.GetCreatedCount))

	return mux
}

// Handles creating a project.
//
// Expects a JSON body of ProjectCreateRequest DTO.
//
// Method: POST
//
// Possible Response Codes:
//   - 400 BadRequest – Invalid input or missing fields
//   - 409 Conflict - Already exists
//   - 500 InternalServerError - Server error
//   - 201 OK - Created successfully
func (prjHndl *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {

	var createRequest dto.ProjectCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ok := createRequest.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := prjHndl.prjSrv.CreateProject(r.Context(), &createRequest); err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Handles creating a project task.
//
// Expects a JSON body of TaskCreateRequest DTO.
//
// Method: POST
//
// Possible Response Codes:
//   - 400 BadRequest – Invalid input, missing fields or constraint violations
//   - 409 Conflict - Already exists
//   - 500 InternalServerError - Server error
//   - 201 OK - Created successfully
func (prjHndl *ProjectHandler) CreateTask(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var createRequest dto.TaskCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ok := createRequest.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := prjHndl.prjSrv.CreateTask(r.Context(), parsedId, &createRequest); err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		if errors.Is(err, service.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Handles get a project with stats.
//
// Optionally expects query parameters:
//   - q: a keyword to search
//   - status: filter by status of a project
//   - page: page number to get
//   - limit: number of project per page
//
// Method: GET
//
// Possible Response Codes:
//   - 400 BadRequest – Invalid query params
//   - 500 InternalServerError - Server error
//   - 200 OK - Request successful
func (prjHndl *ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {

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

	response, err := prjHndl.prjSrv.GetProjects(r.Context(), &dto.ProjectSearchQuery{
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
	json.NewEncoder(w).Encode(response)
}

func (prjHndl *ProjectHandler) GetTasksByProject(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	keyword := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")
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

	resp, err := prjHndl.prjSrv.GetTasksByProject(r.Context(), parsedPrjId, &dto.TaskSearchQuery{
		Keyword:  keyword,
		Status:   status,
		Priority: priority,
		Page:     page,
		Limit:    limit,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (prjHndl *ProjectHandler) GetOneById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := prjHndl.prjSrv.GetOneById(r.Context(), parsedPrjId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (prjHndl *ProjectHandler) GetTaskCompleteCountByProject(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := prjHndl.prjSrv.GetTaskCompleteCountByProjectId(r.Context(), parsedPrjId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (prjHndl *ProjectHandler) GetCreatedCount(w http.ResponseWriter, r *http.Request) {

	resp, err := prjHndl.prjSrv.GetCreatedCount(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (prjHndl *ProjectHandler) SetStatusStarted(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = prjHndl.prjSrv.SetStatusStarted(r.Context(), parsedPrjId)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (prjHndl *ProjectHandler) SetStatusPaused(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = prjHndl.prjSrv.SetStatusPaused(r.Context(), parsedPrjId)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
