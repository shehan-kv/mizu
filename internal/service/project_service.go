package service

import (
	"context"
	"errors"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"mizu/internal/dto/common"
	dto "mizu/internal/dto/project"
	"mizu/internal/event"
	"mizu/internal/logger"
	"mizu/internal/middleware"
)

// Handles project related operations.
type ProjectService struct {
	lg    logger.Logger
	prjSt store.ProjectStore
}

// Creates a new instance of ProjectService.
// It takes a logger, a ProjectStore for project-related data operations.
//
// Parameters:
//   - lg: logger that implements the logger.Logger interface
//   - prjSt: project store that implements the ProjectStore interface
//
// Returns:
//   - a pointer to a new ProjectService
func NewProjectService(lg logger.Logger, prjSt store.ProjectStore) *ProjectService {
	return &ProjectService{
		lg:    lg,
		prjSt: prjSt,
	}
}

// Creates a new project and assigns current signed-in user
// to the project
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//   - request: a pointer to ProjectCreateRequest DTO.
//
// Returns:
//   - ErrAlreadyExists: if project already exists in database.
//   - ErrInternalError: if internal errors occur.
func (prjSrv *ProjectService) CreateProject(ctx context.Context, request *dto.ProjectCreateRequest) error {

	cid := middleware.GetCorrelationID(ctx)

	signedInUser, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		prjSrv.lg.Warn("getting signed in user from context failed",
			"event", event.EventInternalError,
			"correlation_id", cid,
			"scope", "project_service",
			"err", err)
		return ErrInternalError
	}

	_, err = prjSrv.prjSt.CreateOne(ctx, &params.ProjectCreate{
		Name:    request.Name,
		Status:  request.Status,
		Members: append(request.Members, signedInUser.Id)})

	if err != nil {
		if errors.Is(err, store.ErrUniqueViolation) {
			prjSrv.lg.Warn("project with the same name exists",
				"event", event.EventAlreadyExists,
				"correlation_id", cid,
				"project_name", request.Name,
				"scope", "project_service",
				"err", err)
			return ErrAlreadyExists
		}

		prjSrv.lg.Warn("failed to create project",
			"event", event.EventCreateFailed,
			"correlation_id", cid,
			"project_name", request.Name,
			"scope", "project_service",
			"err", err)
		return ErrInternalError
	}

	prjSrv.lg.Info("created project successfully",
		"event", event.EventCreateSuccess,
		"correlation_id", cid,
		"project_name", request.Name,
		"scope", "project_service")

	return nil
}

// Creates a new project task with assigned users.
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//   - request: a pointer to TaskCreateRequest DTO.
//
// Returns:
//   - ErrAlreadyExists: if task already exists in database.
//   - ErrBadRequest: if request parameter violates constraints (eg:- a task for a project that doesn't exist).
//   - ErrInternalError: if internal errors occur.
func (prjSrv *ProjectService) CreateTask(ctx context.Context, projectId int64, request *dto.TaskCreateRequest) error {

	cid := middleware.GetCorrelationID(ctx)

	_, err := prjSrv.prjSt.CreateTask(ctx, &params.TaskCreate{
		ProjectId:            projectId,
		Priority:             request.Priority,
		Status:               request.Status,
		Name:                 request.Name,
		Description:          request.Description,
		EstimatedTimeMinutes: request.EstimatedTimeMinutes,
		Assignees:            request.Assignees})

	if err != nil {
		if errors.Is(err, store.ErrUniqueViolation) {
			prjSrv.lg.Warn("task with the same name exists in the project",
				"event", event.EventAlreadyExists,
				"correlation_id", cid,
				"task_name", request.Name,
				"scope", "project_service",
				"err", err)
			return ErrAlreadyExists
		}

		if errors.Is(err, store.ErrForeignKeyViolation) {
			prjSrv.lg.Warn("task foreign key constraint violated",
				"event", event.EventCreateFailed,
				"correlation_id", cid,
				"task_name", request.Name,
				"scope", "project_service",
				"err", err)
			return ErrBadRequest
		}

		if errors.Is(err, store.ErrNotNullViolation) {
			prjSrv.lg.Warn("task not-null constraint violated",
				"event", event.EventCreateFailed,
				"correlation_id", cid,
				"task_name", request.Name,
				"scope", "project_service",
				"err", err)
			return ErrBadRequest
		}

		prjSrv.lg.Warn("failed to create project task",
			"event", event.EventCreateFailed,
			"correlation_id", cid,
			"task_name", request.Name,
			"scope", "project_service",
			"err", err)
		return ErrInternalError
	}

	prjSrv.lg.Info("created task successfully",
		"event", event.EventCreateSuccess,
		"correlation_id", cid,
		"task_name", request.Name,
		"scope", "project_service")

	return nil
}

// Gets a paginated list of available projects assigned to
// the current user that meets the specified search query.
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//   - query: a pointer to ProjectSearchQuery DTO.
//
// Returns:
//   - ErrInternalError: if internal errors occur.
func (prjSrv *ProjectService) GetProjects(ctx context.Context,
	query *dto.ProjectSearchQuery) (*common.Page[[]dto.ProjectsStatsResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	user, err := middleware.GetUserFromContext(ctx)

	if err != nil {
		prjSrv.lg.Error("could not get user from context",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	projects, err := prjSrv.prjSt.GetWithStats(ctx, &params.ProjectsSearch{
		Keyword: query.Keyword,
		Status:  query.Status,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
		UserId:  user.Id,
	})

	if err != nil {
		prjSrv.lg.Error("could not get projects list",
			"event", event.EventGetFailed,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", user.Id,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	projectResponses := []dto.ProjectsStatsResponse{}
	for _, response := range projects.Items {
		stat := dto.ProjectsStatsResponse{
			Id:             response.Id,
			Name:           response.Name,
			Status:         response.Status,
			CreatedAt:      response.CreatedAt,
			TotalTasks:     response.TotalTasks,
			TasksCompleted: response.TasksCompleted,
			TotalInvoices:  response.TotalInvoices,
			InvoicesPaid:   response.InvoicesPaid,
			TotalQuotes:    response.TotalQuotes,
		}

		projectResponses = append(projectResponses, stat)
	}

	response := &common.Page[[]dto.ProjectsStatsResponse]{
		Count: projects.Total,
		Limit: query.Limit,
		Page:  query.Page,
		Data:  projectResponses,
	}

	return response, nil
}
