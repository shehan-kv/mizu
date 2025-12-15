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
	srv *service.ProjectService
}

// Creates a new instance of ProjectHandler
//
// Parameters:
//   - srv: a pointer to a ProjectService
//
// Returns:
//   - a pointer to a new ProjectHandler
func NewProjectHandler(srv *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{srv: srv}
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
func (h *ProjectHandler) GetMux(
	lg logger.Logger,
	seSt session.SessionStore,
	usrSt store.UserStore) *http.ServeMux {

	mwChain := middleware.NewChain()
	mwChain.Add(
		middleware.CorrelationId(lg),
		middleware.Authenticated(lg, seSt, usrSt))

	mux := http.NewServeMux()

	mux.Handle("POST /", mwChain.Handle(h.CreateProject))
	mux.Handle("GET /", mwChain.Handle(h.GetProjects))
	mux.Handle("GET /{projectId}", mwChain.Handle(h.GetOneById))
	mux.Handle("GET /{projectId}/members", mwChain.Handle(h.GetMembers))
	mux.Handle("PUT /{projectId}/members", mwChain.Handle(h.SetMembers))
	mux.Handle("DELETE /{projectId}", mwChain.Handle(h.DeleteById))
	mux.Handle("POST /{projectId}/tasks", mwChain.Handle(h.CreateTask))
	mux.Handle("GET /{projectId}/tasks", mwChain.Handle(h.GetTasksByProject))
	mux.Handle("GET /{projectId}/tasks/metrics/complete", mwChain.Handle(h.GetTaskCompleteCountByProject))
	mux.Handle("GET /{projectId}/tasks/{taskId}/assignees", mwChain.Handle(h.GetTaskAssignees))
	mux.Handle("PUT /{projectId}/tasks/{taskId}/assignees", mwChain.Handle(h.SetTaskAssignees))
	mux.Handle("POST /{projectId}/status/started", mwChain.Handle(h.SetStatusStarted))
	mux.Handle("POST /{projectId}/status/paused", mwChain.Handle(h.SetStatusPaused))
	mux.Handle("POST /{projectId}/status/cancelled", mwChain.Handle(h.SetStatusCancelled))
	mux.Handle("POST /{projectId}/status/completed", mwChain.Handle(h.SetStatusCompleted))
	mux.Handle("GET /metrics/create", mwChain.Handle(h.GetCreatedCount))

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
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {

	var createRequest dto.ProjectCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ok := createRequest.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.srv.CreateProject(r.Context(), &createRequest); err != nil {
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
func (h *ProjectHandler) CreateTask(w http.ResponseWriter, r *http.Request) {

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

	if err := h.srv.CreateTask(r.Context(), parsedId, &createRequest); err != nil {
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
func (h *ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {

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

	response, err := h.srv.GetProjects(r.Context(), &dto.ProjectSearchQuery{
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

func (h *ProjectHandler) GetTasksByProject(w http.ResponseWriter, r *http.Request) {

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

	resp, err := h.srv.GetTasksByProject(r.Context(), parsedPrjId, &dto.TaskSearchQuery{
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

func (h *ProjectHandler) GetOneById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := h.srv.GetOneById(r.Context(), parsedPrjId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ProjectHandler) GetTaskCompleteCountByProject(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := h.srv.GetTaskCompleteCountByProjectId(r.Context(), parsedPrjId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ProjectHandler) GetCreatedCount(w http.ResponseWriter, r *http.Request) {

	resp, err := h.srv.GetCreatedCount(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ProjectHandler) SetStatusStarted(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.srv.SetStatusStarted(r.Context(), parsedPrjId)
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

func (h *ProjectHandler) SetStatusPaused(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.srv.SetStatusPaused(r.Context(), parsedPrjId)
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

func (h *ProjectHandler) SetStatusCancelled(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.srv.SetStatusCancelled(r.Context(), parsedPrjId)
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

func (h *ProjectHandler) SetStatusCompleted(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.srv.SetStatusCompleted(r.Context(), parsedPrjId)
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

func (h *ProjectHandler) DeleteById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.srv.DeleteById(r.Context(), parsedPrjId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) GetMembers(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := h.srv.GetMembers(r.Context(), parsedPrjId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ProjectHandler) SetMembers(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var setRequest dto.MemberSetRequest

	if err := json.NewDecoder(r.Body).Decode(&setRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ok := setRequest.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.srv.SetMembers(r.Context(), parsedPrjId, &setRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProjectHandler) GetTaskAssignees(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("taskId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := h.srv.GetTaskAssignees(r.Context(), parsedId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ProjectHandler) SetTaskAssignees(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("taskId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var setRequest dto.TaskAssigneeSetRequest

	if err := json.NewDecoder(r.Body).Decode(&setRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ok := setRequest.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.srv.SetTaskAssignees(r.Context(), parsedId, &setRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
