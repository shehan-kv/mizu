package task

import (
	"encoding/json"
	"errors"
	"mizu/internal/application/logger"
	"mizu/internal/application/task"
	domainproject "mizu/internal/domain/project"
	domaintask "mizu/internal/domain/task"
	"mizu/internal/presentation/http/rest/middleware"
	"mizu/internal/presentation/http/rest/page"
	"mizu/internal/presentation/http/rest/query"
	"mizu/internal/presentation/http/rest/response"
	"net/http"
)

type TaskHandler struct {
	taskSrv *task.Service
	log     logger.Logger
}

func NewTaskHandler(taskSrv *task.Service, log logger.Logger) *TaskHandler {
	return &TaskHandler{taskSrv: taskSrv, log: log}
}

func (h *TaskHandler) NewMux(authMiddleware func(http.Handler) http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("POST /tasks/{projectID}", authMiddleware(http.HandlerFunc(h.CreateTask)))
	mux.Handle("GET /tasks/{projectID}", authMiddleware(http.HandlerFunc(h.ListTasks)))
	mux.Handle("GET /tasks/{taskID}/assignees", authMiddleware(http.HandlerFunc(h.ListAssignees)))
	mux.Handle("PUT /tasks/{taskID}/assignees", authMiddleware(http.HandlerFunc(h.ReplaceAssignees)))
	mux.Handle("GET /tasks/{projectID}/completed-count", authMiddleware(http.HandlerFunc(h.GetCompleteCount)))

	return mux
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.taskSrv.Create(r.Context(), task.CreateTaskParams{
		ActorID:          actorID,
		ProjectID:        r.PathValue("projectID"),
		Priority:         req.Priority,
		Status:           req.Status,
		Name:             req.Name,
		Description:      req.Description,
		EstimatedMinutes: req.EstimatedMinutes,
		AssigneeIDs:      req.AssigneeIDs,
	}); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	p, err := page.FromQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.taskSrv.ListTasks(r.Context(), task.ListTaskParams{
		ActorID:   actorID,
		ProjectID: r.PathValue("projectID"),
		Keyword:   query.ExtractString(r, "q"),
		Status:    query.ExtractString(r, "status"),
		Priority:  query.ExtractString(r, "priority"),
		Limit:     p.Limit,
		Offset:    p.Offset,
	})
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	items := make([]TaskResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toTaskResponse(&result.Items[i])
	}

	response.WriteJSON(w, http.StatusOK, page.PaginatedResponse[TaskResponse]{
		Items:      items,
		TotalCount: result.TotalCount,
		Page:       p.Page,
		Limit:      p.Limit,
	})
}

func (h *TaskHandler) ListAssignees(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.taskSrv.ListAssignees(r.Context(), r.PathValue("taskID"), actorID)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	assignees := make([]AssigneeResponse, len(result))
	for i := range result {
		assignees[i] = toAssigneeResponse(&result[i])
	}

	response.WriteJSON(w, http.StatusOK, assignees)
}

func (h *TaskHandler) ReplaceAssignees(w http.ResponseWriter, r *http.Request) {
	var req ReplaceAssigneesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.taskSrv.ReplaceAssignees(
		r.Context(),
		r.PathValue("taskID"),
		req.AssigneeIDs,
		actorID,
	); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TaskHandler) GetCompleteCount(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.taskSrv.GetCompleteCount(r.Context(), r.PathValue("projectID"), actorID)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	resp := make([]MetricResponse, len(result))
	for i := range result {
		resp[i] = toMetricResponse(&result[i])
	}

	response.WriteJSON(w, http.StatusOK, resp)
}

func (h *TaskHandler) writeServiceError(w http.ResponseWriter, method string, path string, err error) {

	switch {

	// 404
	case errors.Is(err, domaintask.ErrTaskNotFound):
		response.WriteError(w, http.StatusNotFound, "task not found")

	case errors.Is(err, domaintask.ErrTaskAssigneeNotFound):
		response.WriteError(w, http.StatusNotFound, "assignee not found")

	// 400
	case errors.Is(err, domaintask.ErrTaskAssigneeNotProjectMember):
		response.WriteError(w, http.StatusBadRequest, "assignee is not a project member")

	case errors.Is(err, domaintask.ErrTaskNameCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "task name cannot be empty")

	case errors.Is(err, domaintask.ErrTaskMinutesMustBePositive):
		response.WriteError(w, http.StatusBadRequest, "estimated minutes must be positive")

	case errors.Is(err, domaintask.ErrTaskInvalidPriority):
		response.WriteError(w, http.StatusBadRequest, "invalid priority")

	case errors.Is(err, domaintask.ErrTaskInvalidStatus):
		response.WriteError(w, http.StatusBadRequest, "invalid status")

	case errors.Is(err, domaintask.ErrTaskIDCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "task id cannot be empty")

	// 409
	case errors.Is(err, domaintask.ErrTaskConcurrentModification):
		response.WriteError(w, http.StatusConflict, "task was modified by another request")

	case errors.Is(err, domainproject.ErrProjectNotAcceptingTasks):
		response.WriteError(w, http.StatusConflict, "project is not accepting tasks")

	// 403
	case errors.Is(err, domainproject.ErrNotProjectMember):
		response.WriteError(w, http.StatusForbidden, "not a project member")

	default:
		h.log.Error(
			"task handler internal server error",
			"error", err,
			"method", method,
			"path", path,
		)

		response.WriteError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
	}
}

func toAssigneeResponse(a *task.AssigneeDTO) AssigneeResponse {
	return AssigneeResponse{
		ID:        a.ID,
		FirstName: a.FirstName,
		LastName:  a.LastName,
		Title:     a.Title,
		Image:     a.Image,
		Role:      a.Role,
	}
}

func toTaskResponse(t *task.TaskDTO) TaskResponse {
	assignees := make([]AssigneeResponse, len(t.Assignees))
	for i := range t.Assignees {
		assignees[i] = toAssigneeResponse(&t.Assignees[i])
	}
	return TaskResponse{
		ID:               t.ID,
		ProjectID:        t.ProjectID,
		Priority:         t.Priority,
		Status:           t.Status,
		Name:             t.Name,
		Description:      t.Description,
		EstimatedMinutes: t.EstimatedMinutes,
		Assignees:        assignees,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}

func toMetricResponse(m *task.MetricDTO) MetricResponse {
	return MetricResponse{
		Key:   m.Key,
		Value: m.Value,
	}
}
