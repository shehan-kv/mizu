package service

import (
	"context"
	"errors"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
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
//   - ErrInternalError: if internal errors occurs.
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

	_, err = prjSrv.prjSt.CreateOne(ctx, &params.ProjectCreateParams{
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
//   - ErrInternalError: if internal errors occurs.
func (prjSrv *ProjectService) CreateTask(ctx context.Context, projectId int64, request *dto.TaskCreateRequest) error {

	cid := middleware.GetCorrelationID(ctx)

	_, err := prjSrv.prjSt.CreateTask(ctx, &params.TaskCreateParams{
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
