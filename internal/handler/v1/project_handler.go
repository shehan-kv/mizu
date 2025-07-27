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
//   - prjSrv: a pointer to a ProjectHandler
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
	mux.Handle("POST /{projectId}/task", mwChain.Handle(prjHndl.CreateTask))

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
//   - 200 OK - Created successfully
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

	w.WriteHeader(http.StatusOK)
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
//   - 200 OK - Created successfully
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

	w.WriteHeader(http.StatusOK)
}
