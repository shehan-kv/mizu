package project

import (
	"encoding/json"
	"errors"
	"mizu/internal/application/logger"
	"mizu/internal/application/project"
	domainiam "mizu/internal/domain/iam"
	domainproject "mizu/internal/domain/project"
	"mizu/internal/presentation/http/rest/middleware"
	"mizu/internal/presentation/http/rest/page"
	"mizu/internal/presentation/http/rest/query"
	"mizu/internal/presentation/http/rest/response"
	"net/http"
)

type ProjectHandler struct {
	projectSrv *project.Service
	log        logger.Logger
}

func NewProjectHandler(projectSrv *project.Service, log logger.Logger) *ProjectHandler {
	return &ProjectHandler{projectSrv: projectSrv, log: log}
}

func (h *ProjectHandler) NewMux(authMiddleware func(http.Handler) http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("POST /projects", authMiddleware(http.HandlerFunc(h.CreateProject)))
	mux.Handle("GET /projects/stats", authMiddleware(http.HandlerFunc(h.ListStats)))
	mux.Handle("GET /projects/stats/members/{memberID}", authMiddleware(http.HandlerFunc(h.ListStatsByMember)))
	mux.Handle("GET /projects/stats/members/{memberID}/all", authMiddleware(http.HandlerFunc(h.ListAllStatsByMember)))
	mux.Handle("GET /projects/stats/created-count", authMiddleware(http.HandlerFunc(h.GetCreatedCount)))
	mux.Handle("GET /projects/{projectID}", authMiddleware(http.HandlerFunc(h.GetProjectOverview)))
	mux.Handle("GET /projects/{projectID}/members", authMiddleware(http.HandlerFunc(h.ListMembers)))
	mux.Handle("PUT /projects/{projectID}/members", authMiddleware(http.HandlerFunc(h.ReplaceMembers)))
	mux.Handle("PUT /projects/{projectID}/start", authMiddleware(http.HandlerFunc(h.StartProject)))
	mux.Handle("PUT /projects/{projectID}/pause", authMiddleware(http.HandlerFunc(h.PauseProject)))
	mux.Handle("PUT /projects/{projectID}/cancel", authMiddleware(http.HandlerFunc(h.CancelProject)))
	mux.Handle("PUT /projects/{projectID}/complete", authMiddleware(http.HandlerFunc(h.CompleteProject)))
	mux.Handle("PUT /projects/members/{memberID}/replace", authMiddleware(http.HandlerFunc(h.ReplaceMemberProjects)))
	mux.Handle("DELETE /projects/{projectID}", authMiddleware(http.HandlerFunc(h.DeleteProject)))

	return mux
}

func (h *ProjectHandler) ReplaceMemberProjects(w http.ResponseWriter, r *http.Request) {
	var req ReplaceMemberProjectsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.projectSrv.ReplaceMemberProjects(
		r.Context(),
		r.PathValue("memberID"),
		actorID, req.ProjectIDs,
	); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.projectSrv.CreateProject(r.Context(), project.CreateProjectParams{
		ActorID: actorID,
		Name:    req.Name,
		Status:  req.Status,
		Members: req.Members,
	}); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *ProjectHandler) ListStats(w http.ResponseWriter, r *http.Request) {
	p, err := page.FromQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.projectSrv.ListStats(r.Context(), project.ListStatsParams{
		ActorID:  actorID,
		MemberID: actorID,
		Keyword:  query.ExtractString(r, "q"),
		Status:   query.ExtractString(r, "status"),
		Limit:    p.Limit,
		Offset:   p.Offset,
	})
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	items := make([]StatsResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toStatsResponse(&result.Items[i])
	}

	response.WriteJSON(w, http.StatusOK, page.PaginatedResponse[StatsResponse]{
		Items:      items,
		TotalCount: result.TotalCount,
		Page:       p.Page,
		Limit:      p.Limit,
	})
}

func (h *ProjectHandler) ListStatsByMember(w http.ResponseWriter, r *http.Request) {
	p, err := page.FromQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.projectSrv.ListStats(r.Context(), project.ListStatsParams{
		ActorID:  actorID,
		MemberID: r.PathValue("memberID"),
		Keyword:  query.ExtractString(r, "q"),
		Status:   query.ExtractString(r, "status"),
		Limit:    p.Limit,
		Offset:   p.Offset,
	})
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	items := make([]StatsResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toStatsResponse(&result.Items[i])
	}

	response.WriteJSON(w, http.StatusOK, page.PaginatedResponse[StatsResponse]{
		Items:      items,
		TotalCount: result.TotalCount,
		Page:       p.Page,
		Limit:      p.Limit,
	})
}

func (h *ProjectHandler) ListAllStatsByMember(w http.ResponseWriter, r *http.Request) {

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.projectSrv.ListAllStats(r.Context(), r.PathValue("memberID"), actorID)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	items := make([]StatsResponse, len(result))
	for i := range result {
		items[i] = toStatsResponse(&result[i])
	}

	response.WriteJSON(w, http.StatusOK, items)
}

func (h *ProjectHandler) GetProjectOverview(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.projectSrv.GetProjectOverview(r.Context(), r.PathValue("projectID"), actorID)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toProjectOverviewResponse(&result))
}

func (h *ProjectHandler) GetCreatedCount(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.projectSrv.GetCreatedCount(r.Context(), actorID)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	metrics := make([]MetricResponse, len(result))
	for i := range result {
		metrics[i] = toMetricResponse(&result[i])
	}

	response.WriteJSON(w, http.StatusOK, metrics)
}

func (h *ProjectHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.projectSrv.ListMembers(r.Context(), project.ListMembersParams{
		ActorID:   actorID,
		ProjectID: r.PathValue("projectID"),
		Keyword:   query.ExtractString(r, "q"),
	})

	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	members := make([]MemberResponse, len(result))
	for i := range result {
		members[i] = toMemberResponse(&result[i])
	}

	response.WriteJSON(w, http.StatusOK, members)
}

func (h *ProjectHandler) ReplaceMembers(w http.ResponseWriter, r *http.Request) {
	var req ReplaceMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.projectSrv.ReplaceMembers(
		r.Context(),
		r.PathValue("projectID"),
		req.MemberIDs,
		actorID,
	); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProjectHandler) StartProject(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.projectSrv.StartProject(r.Context(), r.PathValue("projectID"), actorID); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProjectHandler) PauseProject(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.projectSrv.PauseProject(r.Context(), r.PathValue("projectID"), actorID); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProjectHandler) CancelProject(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.projectSrv.CancelProject(r.Context(), r.PathValue("projectID"), actorID); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProjectHandler) CompleteProject(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.projectSrv.CompleteProject(r.Context(), r.PathValue("projectID"), actorID); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.projectSrv.Delete(r.Context(), r.PathValue("projectID"), actorID); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) writeServiceError(w http.ResponseWriter, method string, path string, err error) {

	switch {

	// 404
	case errors.Is(err, domainproject.ErrProjectNotFound):
		response.WriteError(w, http.StatusNotFound, "project not found")

	case errors.Is(err, domainiam.ErrUserNotFound):
		response.WriteError(w, http.StatusNotFound, "user not found")

	// 403
	case errors.Is(err, domainproject.ErrNotProjectMember):
		response.WriteError(w, http.StatusForbidden, "not a project member")

	// 409
	case errors.Is(err, domainproject.ErrProjectNameAlreadyExists):
		response.WriteError(w, http.StatusConflict, "a project with this name already exists")

	case errors.Is(err, domainproject.ErrProjectNotAcceptingTasks):
		response.WriteError(w, http.StatusConflict, "project is not accepting tasks")

	case errors.Is(err, domainproject.ErrProjectConcurrentModification):
		response.WriteError(w, http.StatusConflict, "project was modified by another request")

	// 400
	case errors.Is(err, domainproject.ErrProjectMustHaveAdministrator):
		response.WriteError(w, http.StatusBadRequest, "project must have at least one administrator")

	case errors.Is(err, domainproject.ErrProjectMustHaveAtLeastOneMember):
		response.WriteError(w, http.StatusBadRequest, "project must have at least one member")

	case errors.Is(err, domainproject.ErrProjectCannotRemoveLastMember):
		response.WriteError(w, http.StatusBadRequest, "cannot remove the last project member")

	case errors.Is(err, domainproject.ErrProjectCannotRemoveUnassignedMember):
		response.WriteError(w, http.StatusBadRequest, "member is not assigned to this project")

	case errors.Is(err, domainproject.ErrProjectNameCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "project name cannot be empty")

	case errors.Is(err, domainproject.ErrProjectIDCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "project id cannot be empty")

	case errors.Is(err, domainproject.ErrProjectMemberIDCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "member id cannot be empty")

	case errors.Is(err, domainproject.ErrProjectInvalidStatus):
		response.WriteError(w, http.StatusBadRequest, "invalid project status")

	case errors.Is(err, domainproject.ErrProjectInvalidMember):
		response.WriteError(w, http.StatusBadRequest, "invalid project member")

	case errors.Is(err, domainproject.ErrProjectCreatorIsRequired):
		response.WriteError(w, http.StatusBadRequest, "project creator is required")

	default:
		h.log.Error(
			"internal server error",
			"method", method,
			"path", path,
			"error", err,
		)

		response.WriteError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
	}

}

func toMemberResponse(m *project.MemberDTO) MemberResponse {
	return MemberResponse{
		ID:        m.ID,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		Title:     m.Title,
		HasImage:  m.HasImage,
		Role:      m.Role,
	}
}

func toMetricResponse(m *project.MetricDTO) MetricResponse {
	return MetricResponse{
		Key:   m.Key,
		Value: m.Value,
	}
}

func toStatsResponse(s *project.StatsDTO) StatsResponse {
	return StatsResponse{
		ID:             s.ID,
		Name:           s.Name,
		Status:         s.Status,
		CreatedAt:      s.CreatedAt,
		TotalTasks:     s.TotalTasks,
		TasksCompleted: s.TasksCompleted,
		TotalInvoices:  s.TotalInvoices,
		InvoicesPaid:   s.InvoicesPaid,
		TotalQuotes:    s.TotalQuotes,
	}
}

func toProjectOverviewResponse(p *project.ProjectOverviewDTO) ProjectOverviewResponse {
	members := make([]MemberResponse, len(p.Members))
	for i := range p.Members {
		members[i] = toMemberResponse(&p.Members[i])
	}
	return ProjectOverviewResponse{
		ID:                  p.ID,
		Name:                p.Name,
		Status:              p.Status,
		CreatedAt:           p.CreatedAt,
		Members:             members,
		TaskCount:           p.TaskCount,
		TaskCompletedCount:  p.TaskCompletedCount,
		InvoiceCount:        p.InvoiceCount,
		InvoicePaidCount:    p.InvoicePaidCount,
		QuoteCount:          p.QuoteCount,
		ContractCount:       p.ContractCount,
		ContractSignedCount: p.ContractSignedCount,
		FileCount:           p.FileCount,
	}
}
